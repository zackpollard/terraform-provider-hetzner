// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http/httptest"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ provider.Provider = &batch3TestProvider{}

type batch3TestProvider struct {
	client *client.Client
}

func newBatch3TestProvider(c *client.Client) func() provider.Provider {
	return func() provider.Provider {
		return &batch3TestProvider{client: c}
	}
}

func batch3ProviderFactories(ts *httptest.Server) map[string]func() (tfprotov6.ProviderServer, error) {
	c := client.NewClient("test", "test")
	c.BaseURL = ts.URL
	return map[string]func() (tfprotov6.ProviderServer, error){
		"hetzner": providerserver.NewProtocol6WithError(newBatch3TestProvider(c)()),
	}
}

func (p *batch3TestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hetzner"
	resp.Version = "test"
}

func (p *batch3TestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{Optional: true, Sensitive: true},
			"password": schema.StringAttribute{Optional: true, Sensitive: true},
		},
	}
}

func (p *batch3TestProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.DataSourceData = p.client
	resp.ResourceData = p.client
}

func (p *batch3TestProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBootRescueResource,
		NewBootLinuxResource,
		NewBootVNCResource,
		NewBootWindowsResource,
	}
}

func (p *batch3TestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBootRescueDataSource,
		NewBootLinuxDataSource,
		NewBootVNCDataSource,
		NewBootWindowsDataSource,
		NewResetDataSource,
		NewWOLDataSource,
	}
}
