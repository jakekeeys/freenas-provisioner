package provisioner

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/extent"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/target"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/target_group"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/target_to_extent"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/storage/dataset"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/storage/z_vol"
	"github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

type Freenas struct {
	Kubernetes kubernetes.Interface
	Freenas    freenas.Interface
	Config     *Config
}

type Config struct {
	RootDatasetName  string
	PortalGroup      int
	InitiatorGroup   int
	ThinProvisioning bool
	ExtentType       string
	LunID            int
	TargetPortal     string
	InitiatorName    string
	ISCSIInterface   string
	FsType           string
}

const (
	// annotation keys
	extentIDAnnotation    = "extentID"
	targetIDAnnotation    = "targetID"
	datasetPoolAnnotation = "datasetPool"
	zVolNameAnnotation    = "zVolName"
)

func (p *Freenas) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	pvName := options.PVName
	pvNamespace := options.PVC.GetObjectMeta().GetNamespace()

	globalConfig, err := p.Freenas.ISCSI().GlobalConfiguration().Get()
	if err != nil {
		return nil, errors.Wrap(err, "error getting global iscsi config")
	}

	rootDs, err := p.Freenas.Storage().Dataset().Get(&dataset.Dataset{Name: &p.Config.RootDatasetName})
	if err != nil {
		return nil, errors.Wrap(err, "error getting root dataset")
	}

	// create zvol
	volSize := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	zVolSize := fmt.Sprintf("%d KiB", int(volSize.Value())/1024)
	zVolName := strings.TrimPrefix(fmt.Sprintf("%s/%s", *rootDs.Name, pvName), *rootDs.Pool+"/")
	zVol, err := p.Freenas.Storage().ZVol().Create(rootDs, &z_vol.ZVol{
		Name:    &zVolName,
		Volsize: &zVolSize,
		Sparse:  &p.Config.ThinProvisioning,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating zvol")
	}

	// create target
	tgt, err := p.Freenas.ISCSI().Target().Create(&target.Target{
		IscsiTargetName: &pvName,
	})
	if err != nil {
		if rollbackErr := p.Freenas.Storage().ZVol().Delete(rootDs, zVol); rollbackErr != nil {
			glog.Warning("error rolling back zvol creation", rollbackErr)
		}
		return nil, errors.Wrap(err, "error creating iscsi target")
	}

	// create target group
	_, err = p.Freenas.ISCSI().TargetGroup().Create(&target_group.TargetGroup{
		IscsiTarget:               tgt.ID,
		IscsiTargetPortalgroup:    &p.Config.PortalGroup,
		IscsiTargetInitiatorgroup: p.Config.InitiatorGroup,
	})
	if err != nil {
		if rollbackErr := p.Freenas.ISCSI().Target().Delete(tgt); rollbackErr != nil {
			glog.Warning("error rolling back iscsi target creation", rollbackErr)
		}
		if rollbackErr := p.Freenas.Storage().ZVol().Delete(rootDs, zVol); rollbackErr != nil {
			glog.Warning("error rolling back zvol creation", rollbackErr)
		}
		return nil, errors.Wrap(err, "error creating iscsi target group")
	}

	// create extent
	extentDisk := fmt.Sprintf("zvol/%s/%s", *rootDs.Pool, *zVol.Name)
	ext, err := p.Freenas.ISCSI().Extent().Create(&extent.Extent{
		IscsiTargetExtentType: &p.Config.ExtentType,
		IscsiTargetExtentName: &pvName,
		IscsiTargetExtentDisk: &extentDisk,
	})
	if err != nil {
		if rollbackErr := p.Freenas.ISCSI().Target().Delete(tgt); rollbackErr != nil {
			glog.Warning("error rolling back iscsi target creation", rollbackErr)
		}
		if rollbackErr := p.Freenas.Storage().ZVol().Delete(rootDs, zVol); rollbackErr != nil {
			glog.Warning("error rolling back zvol creation", rollbackErr)
		}
		return nil, errors.Wrap(err, "error creating iscsi extent")
	}

	// create target to extent
	_, err = p.Freenas.ISCSI().TargetToExtent().Create(&target_to_extent.TargetToExtent{
		IscsiTarget: tgt.ID,
		IscsiExtent: ext.ID,
		IscsiLunid:  p.Config.LunID,
	})
	if err != nil {
		if rollbackErr := p.Freenas.ISCSI().Extent().Delete(ext); rollbackErr != nil {
			glog.Warning("error rolling back iscsi extent creation", rollbackErr)
		}
		if rollbackErr := p.Freenas.ISCSI().Target().Delete(tgt); rollbackErr != nil {
			glog.Warning("error rolling back iscsi target creation", rollbackErr)
		}
		if rollbackErr := p.Freenas.Storage().ZVol().Delete(rootDs, zVol); rollbackErr != nil {
			glog.Warning("error rolling back zvol creation", rollbackErr)
		}
		return nil, errors.Wrap(err, "error creating iscsi target to extent")
	}

	return &v1.PersistentVolume{
		ObjectMeta: v12.ObjectMeta{
			Name:      pvName,
			Namespace: pvNamespace,
			Annotations: map[string]string{
				extentIDAnnotation:    strconv.Itoa(*ext.ID),
				targetIDAnnotation:    strconv.Itoa(*tgt.ID),
				datasetPoolAnnotation: *rootDs.Pool,
				zVolNameAnnotation:    *zVol.Name,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				ISCSI: &v1.ISCSIPersistentVolumeSource{
					TargetPortal:   p.Config.TargetPortal,
					IQN:            fmt.Sprintf("%s:%s", *globalConfig.IscsiBasename, pvName),
					Lun:            int32(p.Config.LunID),
					ISCSIInterface: p.Config.ISCSIInterface,
					FSType:         p.Config.FsType,
					InitiatorName:  &p.Config.InitiatorName,
				},
			},
			AccessModes:                   options.PVC.Spec.AccessModes,
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			StorageClassName:              *options.PVC.Spec.StorageClassName,
			MountOptions:                  options.MountOptions,
			VolumeMode:                    options.PVC.Spec.VolumeMode,
		},
	}, nil
}

func (p *Freenas) Delete(volume *v1.PersistentVolume) error {
	// delete extent
	extentIDString, ok := volume.Annotations[extentIDAnnotation]
	if !ok {
		return fmt.Errorf("missing required volume annotation %s", extentIDAnnotation)
	}

	extentID, err := strconv.Atoi(extentIDString)
	if err != nil {
		return errors.Wrapf(err, "error converting parameter %s", extentIDAnnotation)
	}

	err = p.Freenas.ISCSI().Extent().Delete(&extent.Extent{
		ID: &extentID,
	})
	if err != nil {
		return errors.Wrap(err, "error deleting extent")
	}

	// delete target which also removes associated target groups and target to extents
	targetIDString, ok := volume.Annotations[targetIDAnnotation]
	if !ok {
		return fmt.Errorf("missing required volume annotation %s", targetIDAnnotation)
	}

	targetID, err := strconv.Atoi(targetIDString)
	if err != nil {
		return errors.Wrapf(err, "error converting parameter %s", targetIDAnnotation)
	}

	err = p.Freenas.ISCSI().Target().Delete(&target.Target{
		ID: &targetID,
	})
	if err != nil {
		return errors.Wrap(err, "error deleting target")
	}

	datasetPool, ok := volume.Annotations[datasetPoolAnnotation]
	if !ok {
		return fmt.Errorf("missing required volume annotation %s", datasetPoolAnnotation)
	}

	zVolName, ok := volume.Annotations[zVolNameAnnotation]
	if !ok {
		return fmt.Errorf("missing required volume annotation %s", zVolNameAnnotation)
	}

	// delete zvol
	err = p.Freenas.Storage().ZVol().Delete(
		&dataset.Dataset{
			Pool: &datasetPool,
		},
		&z_vol.ZVol{
			Name: &zVolName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error deleting zvol")
	}

	return nil
}
