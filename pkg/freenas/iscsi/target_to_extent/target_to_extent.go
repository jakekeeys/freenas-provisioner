package target_to_extent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/services/iscsi/targettoextent"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Create(targetToExtent *TargetToExtent) (*TargetToExtent, error)
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type TargetToExtent struct {
	IscsiTarget *int        `json:"iscsi_target,omitempty"`
	IscsiExtent *int        `json:"iscsi_extent,omitempty"`
	IscsiLunid  interface{} `json:"iscsi_lunid,omitempty"`
	ID          *int        `json:"id,omitempty"`
}

func (c Client) Create(targetToExtent *TargetToExtent) (*TargetToExtent, error) {
	targetToExtentBytes, err := json.Marshal(targetToExtent)
	if err != nil {
		return nil, err
	}

	request, err := c.client.NewRequest(http.MethodPost, fmt.Sprintf("%s/", basePath), bytes.NewReader(targetToExtentBytes))
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

	var tte TargetToExtent
	err = json.Unmarshal(body, &tte)
	if err != nil {
		return nil, err
	}

	return &tte, nil
}
