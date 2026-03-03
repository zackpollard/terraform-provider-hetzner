// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ datasource.DataSource              = &firewallTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &firewallTemplateDataSource{}
)

type firewallTemplateDataSource struct {
	client *client.Client
}

type firewallTemplateDataSourceModel struct {
	ID           types.String        `tfsdk:"id"`
	Name         types.String        `tfsdk:"name"`
	FilterIPv6   types.Bool          `tfsdk:"filter_ipv6"`
	AllowlistHOS types.Bool          `tfsdk:"allowlist_hos"`
	IsDefault    types.Bool          `tfsdk:"is_default"`
	InputRules   []firewallRuleModel `tfsdk:"input"`
	OutputRules  []firewallRuleModel `tfsdk:"output"`
}

func NewFirewallTemplateDataSource() datasource.DataSource {
	return &firewallTemplateDataSource{}
}

func (d *firewallTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_template"
}

func (d *firewallTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a firewall template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Template ID.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name.",
				Computed:            true,
			},
			"filter_ipv6": schema.BoolAttribute{
				MarkdownDescription: "Whether template filters IPv6 traffic.",
				Computed:            true,
			},
			"allowlist_hos": schema.BoolAttribute{
				MarkdownDescription: "Allow Hetzner services.",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the default template.",
				Computed:            true,
			},
			"input": schema.ListNestedAttribute{
				MarkdownDescription: "Incoming traffic rules.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: firewallRuleDataSourceSchemaAttrs,
				},
			},
			"output": schema.ListNestedAttribute{
				MarkdownDescription: "Outgoing traffic rules.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: firewallRuleDataSourceSchemaAttrs,
				},
			},
		},
	}
}

func (d *firewallTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *firewallTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data firewallTemplateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get("/firewall/template/" + data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall template", err.Error())
		return
	}

	var apiResp firewallTemplateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall template response", err.Error())
		return
	}

	tmpl := apiResp.FirewallTemplate
	data.ID = types.StringValue(fmt.Sprintf("%d", tmpl.ID))
	data.Name = types.StringValue(tmpl.Name)
	data.FilterIPv6 = types.BoolValue(tmpl.FilterIPv6)
	data.AllowlistHOS = types.BoolValue(tmpl.AllowlistHOS)
	data.IsDefault = types.BoolValue(tmpl.IsDefault)
	data.InputRules = apiRulesToModel(tmpl.Rules.Input)
	data.OutputRules = apiRulesToModel(tmpl.Rules.Output)

	if data.InputRules == nil {
		data.InputRules = []firewallRuleModel{}
	}
	if data.OutputRules == nil {
		data.OutputRules = []firewallRuleModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
