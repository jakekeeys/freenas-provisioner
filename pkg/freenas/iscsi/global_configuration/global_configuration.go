package global_configuration

import (
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/services/iscsi/globalconfiguration"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Get() (*GlobalConfiguration, error)
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type GlobalConfiguration struct {
	IscsiBasename           *string     `json:"iscsi_basename,omitempty"`
	IscsiIsnsServers        *string     `json:"iscsi_isns_servers,omitempty"`
	IscsiPoolAvailThreshold interface{} `json:"iscsi_pool_avail_threshold,omitempty"`
	ID                      *int        `json:"id,omitempty"`
}

func (c Client) Get() (*GlobalConfiguration, error) {
	request, err := c.client.NewRequest(http.MethodGet, fmt.Sprintf("%s/", basePath), nil)
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

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s, body: %s", response.Status, string(body))
	}

	var gc GlobalConfiguration
	err = json.Unmarshal(body, &gc)
	if err != nil {
		return nil, err
	}

	return &gc, nil
}
