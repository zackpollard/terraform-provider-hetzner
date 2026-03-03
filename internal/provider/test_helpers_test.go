// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ provider.Provider = &batch1TestProvider{}

// batch1TestProvider is a test provider for batch 1 resources (SSH key, rDNS, firewall, firewall template).
type batch1TestProvider struct {
	client *client.Client
}

func newBatch1TestProvider(c *client.Client) func() provider.Provider {
	return func() provider.Provider {
		return &batch1TestProvider{client: c}
	}
}

func batch1ProviderFactories(c *client.Client) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"hetzner": providerserver.NewProtocol6WithError(newBatch1TestProvider(c)()),
	}
}

func (p *batch1TestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hetzner"
	resp.Version = "test"
}

func (p *batch1TestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{Optional: true, Sensitive: true},
			"password": schema.StringAttribute{Optional: true, Sensitive: true},
		},
	}
}

func (p *batch1TestProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.DataSourceData = p.client
	resp.ResourceData = p.client
}

func (p *batch1TestProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSSHKeyResource,
		NewRDNSResource,
		NewFirewallResource,
		NewFirewallTemplateResource,
	}
}

func (p *batch1TestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSSHKeyDataSource,
		NewSSHKeysDataSource,
		NewRDNSDataSource,
		NewFirewallDataSource,
		NewFirewallTemplateDataSource,
		NewFirewallTemplatesDataSource,
	}
}
