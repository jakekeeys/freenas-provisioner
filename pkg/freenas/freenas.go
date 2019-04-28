package freenas

import (
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/storage"
)

type Client struct {
	client  rest.Interface
	iscsi   iscsi.Interface
	storage storage.Interface
}

type Interface interface {
	ISCSI() iscsi.Interface
	Storage() storage.Interface
}

func New(client rest.Interface) Interface {
	return &Client{
		client:  client,
		iscsi:   iscsi.New(client),
		storage: storage.New(client),
	}
}

func (f Client) ISCSI() iscsi.Interface {
	return f.iscsi
}

func (f Client) Storage() storage.Interface {
	return f.storage
}
