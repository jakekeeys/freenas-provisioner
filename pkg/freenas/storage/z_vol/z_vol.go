package z_vol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/storage/dataset"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/storage/volume"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Create(dataset *dataset.Dataset, zVol *ZVol) (*ZVol, error)
	Delete(dataset *dataset.Dataset, zVol *ZVol) error
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type ZVol struct {
	Comments    *string     `json:"comments,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Volsize     interface{} `json:"volsize,omitempty"`
	Compression *string     `json:"compression,omitempty"`
	Sparse      *bool       `json:"sparse,omitempty"`
	Force       *bool       `json:"force,omitempty"`
	Blocksize   *string     `json:"blocksize,omitempty"`
	Avail       *int64      `json:"avail,omitempty"`
	Dedup       *string     `json:"dedup,omitempty"`
	Refer       *int        `json:"refer,omitempty"`
	Used        *int        `json:"used,omitempty"`
}

func (c Client) Delete(dataset *dataset.Dataset, zVol *ZVol) error {
	request, err := c.client.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s/zvols/%s/", basePath, *dataset.Pool, *zVol.Name), nil)
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

func (c Client) Create(dataset *dataset.Dataset, zVol *ZVol) (*ZVol, error) {
	zVolBytes, err := json.Marshal(zVol)
	if err != nil {
		return nil, err
	}

	request, err := c.client.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/zvols/", basePath, *dataset.Pool), bytes.NewReader(zVolBytes))
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

	if response.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status code: %s, body: %s", response.Status, string(body))
	}

	var zv ZVol
	err = json.Unmarshal(body, &zv)
	if err != nil {
		return nil, err
	}

	return &zv, nil
}
