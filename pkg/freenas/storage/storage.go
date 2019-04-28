package storage

import (
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/storage/dataset"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/storage/z_vol"
)

type Client struct {
	client  rest.Interface
	dataset dataset.Interface
	zvol    z_vol.Interface
}

type Interface interface {
	Dataset() dataset.Interface
	ZVol() z_vol.Interface
}

func New(client rest.Interface) Interface {
	return &Client{
		client:  client,
		dataset: dataset.New(client),
		zvol:    z_vol.New(client),
	}
}

func (s Client) Dataset() dataset.Interface {
	return s.dataset
}

func (s Client) ZVol() z_vol.Interface {
	return s.zvol
}
