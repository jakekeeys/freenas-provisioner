package iscsi

import (
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/extent"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/global_configuration"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/target"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/target_group"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/iscsi/target_to_extent"
	"github.com/jakekeeys/freenas-provisioner/pkg/freenas/rest"
)

type Client struct {
	client              rest.Interface
	globalConfiguration global_configuration.Interface
	target              target.Interface
	extent              extent.Interface
	targetToExtent      target_to_extent.Interface
	targetGroup         target_group.Interface
}

type Interface interface {
	GlobalConfiguration() global_configuration.Interface
	Target() target.Interface
	Extent() extent.Interface
	TargetToExtent() target_to_extent.Interface
	TargetGroup() target_group.Interface
}

func New(client rest.Interface) Interface {
	return &Client{
		client:              client,
		globalConfiguration: global_configuration.New(client),
		target:              target.New(client),
		extent:              extent.New(client),
		targetToExtent:      target_to_extent.New(client),
		targetGroup:         target_group.New(client),
	}
}

func (c Client) GlobalConfiguration() global_configuration.Interface {
	return c.globalConfiguration
}

func (c Client) Target() target.Interface {
	return c.target
}

func (c Client) Extent() extent.Interface {
	return c.extent
}

func (c Client) TargetToExtent() target_to_extent.Interface {
	return c.targetToExtent
}

func (c Client) TargetGroup() target_group.Interface {
	return c.targetGroup
}
