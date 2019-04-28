package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/jakekeeys/freenas-provisioner/internal/provisioner"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas"
	freenas_rest "github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"github.com/jawher/mow.cli"
	"github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strconv"
)

const (
	appName = "freenas-provisioner"
	appDesc = "Kubernetes FreeNAS Provisioner"
)

var (
	appVersion string
)

const (
	// parameter keys
	rootDatasetNameParam  = "rootDatasetName"
	portalGroupParam      = "portalGroup"
	initiatorGroupParam   = "initiatorGroup"
	lunIDParam            = "lunID"
	thinProvisioningParam = "thinProvisioning"
	targetPortalParam     = "targetPortal"
	initiatorNameParam    = "initiatorName"

	// parameter defaults
	extentType       = "Disk"
	iSCSIInterface   = "default"
	fsType           = "ext4"
	thinProvisioning = true
	initiatorName    = "iqn.2001-04.com.kubernetes:storage"
)

func main() {
	flag.Parse()
	err := flag.Set("logtostderr", "true")
	if err != nil {
		glog.Fatal(err)
	}

	app := cli.App(appName, appDesc)
	app.Version(appName, appVersion)

	kubenetesConfig := app.String(cli.StringOpt{
		Name:   "kubernetes-config",
		Desc:   "Path to kubernetes configuration file (for out of cluster execution)",
		EnvVar: "KUBECONFIG",
	})
	provisionerName := app.String(cli.StringOpt{
		Name:   "provisioner-name",
		Value:  "freenas-provisoner",
		Desc:   "Provisioner Name (e.g. 'provisioner' attribute of storage-class)",
		EnvVar: "PROVISIONER_NAME",
	})
	storageClassName := app.String(cli.StringOpt{
		Name:   "storage-class-name",
		Value:  "freenas-iscsi",
		Desc:   "Storage class name",
		EnvVar: "STORAGE_CLASS_NAME",
	})

	freenasAPIUser := app.String(cli.StringOpt{
		Name:   "freenas-api-user",
		Value:  "root",
		Desc:   "Freenas API username",
		EnvVar: "FREENAS_API_USER",
	})
	freenasAPIPassword := app.String(cli.StringOpt{
		Name:   "freenas-api-password",
		Desc:   "Freenas API password",
		EnvVar: "FREENAS_API_PASSWORD",
	})
	freenasAPIHost := app.String(cli.StringOpt{
		Name:   "freenas-api-host",
		Desc:   "Freenas API host",
		EnvVar: "FREENAS_API_HOST",
	})
	freenasAPISkipTLSVerification := app.Bool(cli.BoolOpt{
		Name:   "freenas-api-skip-tls-verification",
		Desc:   "Skip tls certificate verification",
		EnvVar: "FREENAS_API_SKIP_TLS_VERIFICATION",
		Value:  false,
	})

	app.Action = func() {
		var config *rest.Config
		var err error
		if *kubenetesConfig != "" {
			config, err = clientcmd.BuildConfigFromFlags("", *kubenetesConfig)
			if err != nil {
				glog.Fatal(err)
			}
		} else {
			config, err = rest.InClusterConfig()
			if err != nil {
				glog.Fatal(err)
			}
		}

		k8sClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			glog.Fatal(err)
		}

		serverVersion, err := k8sClient.Discovery().ServerVersion()
		if err != nil {
			glog.Fatal(err)
		}

		class, err := k8sClient.StorageV1().StorageClasses().Get(*storageClassName, v1.GetOptions{})
		if err != nil {
			glog.Fatal(err)
		}

		freenasProvisionerConfig := provisioner.Config{
			ExtentType:       extentType,
			ISCSIInterface:   iSCSIInterface,
			FsType:           fsType,
			ThinProvisioning: thinProvisioning,
			InitiatorName:    initiatorName,
		}

		// required params
		rootDatasetName, ok := class.Parameters[rootDatasetNameParam]
		if !ok {
			glog.Fatal("missing required storage class parameter", rootDatasetNameParam)
		}
		freenasProvisionerConfig.RootDatasetName = rootDatasetName

		portalGroupString, ok := class.Parameters[portalGroupParam]
		if !ok {
			glog.Fatal("missing required storage class parameter", portalGroupParam)
		}
		portalGroup, err := strconv.Atoi(portalGroupString)
		if err != nil {
			glog.Fatal(err)
		}
		freenasProvisionerConfig.PortalGroup = portalGroup

		initiatorGroupString, ok := class.Parameters[initiatorGroupParam]
		if !ok {
			glog.Fatal("missing required storage class parameter", initiatorGroupParam)
		}
		initiatorGroup, err := strconv.Atoi(initiatorGroupString)
		if err != nil {
			glog.Fatal(err)
		}
		freenasProvisionerConfig.InitiatorGroup = initiatorGroup

		lunIDString, ok := class.Parameters[lunIDParam]
		if !ok {
			glog.Fatal("missing required storage class parameter", lunIDParam)
		}
		lunID, err := strconv.Atoi(lunIDString)
		if err != nil {
			glog.Fatal(err)
		}
		freenasProvisionerConfig.LunID = lunID

		targetPortal, ok := class.Parameters[targetPortalParam]
		if !ok {
			glog.Fatal("missing required storage class parameter", targetPortalParam)
		}
		freenasProvisionerConfig.TargetPortal = targetPortal

		// optional params
		if thinProvisioning, ok := class.Parameters[thinProvisioningParam]; ok {
			thinProvisioning, err := strconv.ParseBool(thinProvisioning)
			if err != nil {
				glog.Fatal(err)
			}
			freenasProvisionerConfig.ThinProvisioning = thinProvisioning
		}

		if initiatorName, ok := class.Parameters[initiatorNameParam]; ok {
			freenasProvisionerConfig.InitiatorName = initiatorName
		}

		fnClient := freenas.New(freenas_rest.New(*freenasAPIUser, *freenasAPIPassword, *freenasAPIHost, *freenasAPISkipTLSVerification))
		freenasProvisioner := &provisioner.Freenas{
			Kubernetes: k8sClient,
			Freenas:    fnClient,
			Config:     &freenasProvisionerConfig,
		}

		pc := controller.NewProvisionController(k8sClient, *provisionerName, freenasProvisioner, serverVersion.GitVersion)
		pc.Run(wait.NeverStop)
	}

	err = app.Run(os.Args)
	if err != nil {
		glog.Fatal(err)
	}
}
