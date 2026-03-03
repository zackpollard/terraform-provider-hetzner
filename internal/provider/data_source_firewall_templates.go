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
	_ datasource.DataSource              = &firewallTemplatesDataSource{}
	_ datasource.DataSourceWithConfigure = &firewallTemplatesDataSource{}
)

type firewallTemplatesDataSource struct {
	client *client.Client
}

type firewallTemplatesDataSourceModel struct {
	Templates []firewallTemplateDataSourceModel `tfsdk:"templates"`
}

func NewFirewallTemplatesDataSource() datasource.DataSource {
	return &firewallTemplatesDataSource{}
}

func (d *firewallTemplatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_templates"
}

func (d *firewallTemplatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to list all firewall templates.",
		Attributes: map[string]schema.Attribute{
			"templates": schema.ListNestedAttribute{
				MarkdownDescription: "List of firewall templates.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Template ID.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *firewallTemplatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *firewallTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/firewall/template")
	if err != nil {
		resp.Diagnostics.AddError("Error listing firewall templates", err.Error())
		return
	}

	var apiTemplates []firewallTemplateAPIResponse
	if err := json.Unmarshal(body, &apiTemplates); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall templates response", err.Error())
		return
	}

	var data firewallTemplatesDataSourceModel
	for _, t := range apiTemplates {
		tmpl := t.FirewallTemplate
		inputRules := apiRulesToModel(tmpl.Rules.Input)
		outputRules := apiRulesToModel(tmpl.Rules.Output)
		if inputRules == nil {
			inputRules = []firewallRuleModel{}
		}
		if outputRules == nil {
			outputRules = []firewallRuleModel{}
		}
		data.Templates = append(data.Templates, firewallTemplateDataSourceModel{
			ID:           types.StringValue(fmt.Sprintf("%d", tmpl.ID)),
			Name:         types.StringValue(tmpl.Name),
			FilterIPv6:   types.BoolValue(tmpl.FilterIPv6),
			AllowlistHOS: types.BoolValue(tmpl.AllowlistHOS),
			IsDefault:    types.BoolValue(tmpl.IsDefault),
			InputRules:   inputRules,
			OutputRules:  outputRules,
		})
	}

	if data.Templates == nil {
		data.Templates = []firewallTemplateDataSourceModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
