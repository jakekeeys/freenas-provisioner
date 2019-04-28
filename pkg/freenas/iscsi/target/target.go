package target

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/services/iscsi/target"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Create(target *Target) (*Target, error)
	Delete(target *Target) error
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type Target struct {
	IscsiTargetName  *string     `json:"iscsi_target_name,omitempty"`
	IscsiTargetAlias interface{} `json:"iscsi_target_alias,omitempty"`
	ID               *int        `json:"id,omitempty"`
}

func (c Client) Delete(target *Target) error {
	request, err := c.client.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d/", basePath, *target.ID), nil)
	if err != nil {
		return err
	}

	response, err := c.client.DoRequest(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %s, body: %s", response.Status, string(body))
	}

	return nil
}

func (c Client) Create(target *Target) (*Target, error) {
	targetBytes, err := json.Marshal(target)
	if err != nil {
		return nil, err
	}

	request, err := c.client.NewRequest(http.MethodPost, fmt.Sprintf("%s/", basePath), bytes.NewReader(targetBytes))
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

	var t Target
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
