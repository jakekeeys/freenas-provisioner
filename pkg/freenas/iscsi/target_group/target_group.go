package target_group

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/services/iscsi/targetgroup"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Create(targetGroup *TargetGroup) (*TargetGroup, error)
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type TargetGroup struct {
	IscsiTarget               *int        `json:"iscsi_target,omitempty"`
	IscsiTargetAuthgroup      interface{} `json:"iscsi_target_authgroup,omitempty"`
	IscsiTargetAuthtype       *string     `json:"iscsi_target_authtype,omitempty"`
	IscsiTargetPortalgroup    *int        `json:"iscsi_target_portalgroup,omitempty"`
	IscsiTargetInitiatorgroup interface{} `json:"iscsi_target_initiatorgroup,omitempty"`
	IscsiTargetInitialdigest  *string     `json:"iscsi_target_initialdigest,omitempty"`
}

func (c Client) Create(targetGroup *TargetGroup) (*TargetGroup, error) {
	targetGroupBytes, err := json.Marshal(targetGroup)
	if err != nil {
		return nil, err
	}

	request, err := c.client.NewRequest(http.MethodPost, fmt.Sprintf("%s/", basePath), bytes.NewReader(targetGroupBytes))
	if err != nil {
		return nil, err
	}

	response, err := c.client.DoRequest(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %s, body: %s", response.Status, string(body))
	}

	var tg TargetGroup
	err = json.Unmarshal(body, &tg)
	if err != nil {
		return nil, err
	}

	return &tg, nil
}
