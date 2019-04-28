package dataset

import (
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/storage/dataset"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Get(dataset *Dataset) (*Dataset, error)
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type Dataset struct {
	Atime          *string       `json:"atime,omitempty"`
	Avail          *int64        `json:"avail,omitempty"`
	Comments       *string       `json:"comments,omitempty"`
	Compression    *string       `json:"compression,omitempty"`
	Dedup          *string       `json:"dedup,omitempty"`
	InheritProps   []interface{} `json:"inherit_props,omitempty"`
	Mountpoint     *string       `json:"mountpoint,omitempty"`
	Name           *string       `json:"name,omitempty"`
	Pool           *string       `json:"pool,omitempty"`
	Quota          *int          `json:"quota,omitempty"`
	Readonly       *string       `json:"readonly,omitempty"`
	Recordsize     *int          `json:"recordsize,omitempty"`
	Refer          *int          `json:"refer,omitempty"`
	Refquota       *int          `json:"refquota,omitempty"`
	Refreservation *int          `json:"refreservation,omitempty"`
	Reservation    *int          `json:"reservation,omitempty"`
	Used           *int64        `json:"used,omitempty"`
}

func (c Client) Get(dataset *Dataset) (*Dataset, error) {
	request, err := c.client.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/", basePath, *dataset.Name), nil)
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

	var ds Dataset
	err = json.Unmarshal(body, &ds)
	if err != nil {
		return nil, err
	}

	return &ds, nil
}
