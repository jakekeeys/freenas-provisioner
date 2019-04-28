package extent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"io/ioutil"
	"net/http"
)

const basePath = "/api/v1.0/services/iscsi/extent"

type Client struct {
	client rest.Interface
}

type Interface interface {
	Create(target *Extent) (*Extent, error)
	Delete(extent *Extent) error
}

func New(client rest.Interface) Interface {
	return &Client{
		client: client,
	}
}

type Extent struct {
	IscsiTargetExtentComment        *string     `json:"iscsi_target_extent_comment,omitempty"`
	IscsiTargetExtentType           *string     `json:"iscsi_target_extent_type,omitempty"`
	IscsiTargetExtentName           *string     `json:"iscsi_target_extent_name,omitempty"`
	IscsiTargetExtentFilesize       *string     `json:"iscsi_target_extent_filesize,omitempty"`
	IscsiTargetExtentInsecureTpc    *bool       `json:"iscsi_target_extent_insecure_tpc,omitempty"`
	IscsiTargetExtentNaa            *string     `json:"iscsi_target_extent_naa,omitempty"`
	ID                              *int        `json:"id,omitempty"`
	IscsiTargetExtentPath           *string     `json:"iscsi_target_extent_path,omitempty"`
	IscsiTargetExtentXen            *bool       `json:"iscsi_target_extent_xen,omitempty"`
	IscsiTargetExtentDisk           *string     `json:"iscsi_target_extent_disk,omitempty"`
	IscsiTargetExtentAvailThreshold interface{} `json:"iscsi_target_extent_avail_threshold,omitempty"`
	IscsiTargetExtentBlocksize      *int        `json:"iscsi_target_extent_blocksize,omitempty"`
	IscsiTargetExtentPblocksize     *bool       `json:"iscsi_target_extent_pblocksize,omitempty"`
	IscsiTargetExtentRpm            *string     `json:"iscsi_target_extent_rpm,omitempty"`
	IscsiTargetExtentRo             *bool       `json:"iscsi_target_extent_ro,omitempty"`
	IscsiTargetExtentSerial         *string     `json:"iscsi_target_extent_serial,omitempty"`
}

func (c Client) Delete(extent *Extent) error {
	request, err := c.client.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d/", basePath, *extent.ID), nil)
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

func (c Client) Create(extent *Extent) (*Extent, error) {
	extentBytes, err := json.Marshal(extent)
	if err != nil {
		return nil, err
	}

	request, err := c.client.NewRequest(http.MethodPost, fmt.Sprintf("%s/", basePath), bytes.NewReader(extentBytes))
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

	var e Extent
	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}
