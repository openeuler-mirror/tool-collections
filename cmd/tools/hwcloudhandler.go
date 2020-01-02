package main

import (
	"context"
	"fmt"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
)

type HWCloudHandler struct {
	Client *golangsdk.ProviderClient
	Name string
	Passwd string
	Context context.Context
}

func NewHWCloudHandler(authUrl, user, password, region string) *HWCloudHandler{
	// oauth
	opts := golangsdk.AuthOptions{
		IdentityEndpoint: authUrl,
		Username: user,
		Password: password,
		DomainName: "openeuler",
		TenantName: region,
	}

	opts, err := openstack.AuthOptionsFromEnv()

	if err != nil {
		fmt.Printf("Failed to initialize hwcloudhandler due to error %v", err)
	}

	provider, err := openstack.AuthenticatedClient(opts)

	if err != nil {
		fmt.Printf("Failed to initialize hwcloudhandler due to error %v", err)
	}

	return &HWCloudHandler{
		Name:user,
		Passwd:password,
		Client: provider,
	}
}
