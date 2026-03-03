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
	_ datasource.DataSource              = &firewallDataSource{}
	_ datasource.DataSourceWithConfigure = &firewallDataSource{}
)

type firewallDataSource struct {
	client *client.Client
}

type firewallDataSourceModel struct {
	ServerNumber types.String        `tfsdk:"server_number"`
	ServerIP     types.String        `tfsdk:"server_ip"`
	Status       types.String        `tfsdk:"status"`
	AllowlistHOS types.Bool          `tfsdk:"allowlist_hos"`
	FilterIPv6   types.Bool          `tfsdk:"filter_ipv6"`
	Port         types.String        `tfsdk:"port"`
	InputRules   []firewallRuleModel `tfsdk:"input"`
	OutputRules  []firewallRuleModel `tfsdk:"output"`
}

func NewFirewallDataSource() datasource.DataSource {
	return &firewallDataSource{}
}

func (d *firewallDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

var firewallRuleDataSourceSchemaAttrs = map[string]schema.Attribute{
	"ip_version": schema.StringAttribute{
		MarkdownDescription: "IP version.",
		Computed:            true,
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Rule name.",
		Computed:            true,
	},
	"dst_ip": schema.StringAttribute{
		MarkdownDescription: "Destination IP/subnet in CIDR notation.",
		Computed:            true,
	},
	"src_ip": schema.StringAttribute{
		MarkdownDescription: "Source IP/subnet in CIDR notation.",
		Computed:            true,
	},
	"dst_port": schema.StringAttribute{
		MarkdownDescription: "Destination port or range.",
		Computed:            true,
	},
	"src_port": schema.StringAttribute{
		MarkdownDescription: "Source port or range.",
		Computed:            true,
	},
	"protocol": schema.StringAttribute{
		MarkdownDescription: "Protocol.",
		Computed:            true,
	},
	"tcp_flags": schema.StringAttribute{
		MarkdownDescription: "TCP flags.",
		Computed:            true,
	},
	"action": schema.StringAttribute{
		MarkdownDescription: "Action: accept or discard.",
		Computed:            true,
	},
}

func (d *firewallDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read firewall rules for a Hetzner dedicated server.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.StringAttribute{
				MarkdownDescription: "Server number.",
				Required:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Server main IP address.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Firewall status.",
				Computed:            true,
			},
			"allowlist_hos": schema.BoolAttribute{
				MarkdownDescription: "Allow Hetzner services.",
				Computed:            true,
			},
			"filter_ipv6": schema.BoolAttribute{
				MarkdownDescription: "Whether firewall also filters IPv6.",
				Computed:            true,
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Switch port.",
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

func (d *firewallDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *firewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data firewallDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get("/firewall/" + data.ServerNumber.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall", err.Error())
		return
	}

	var apiResp firewallAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall response", err.Error())
		return
	}

	fw := apiResp.Firewall
	data.ServerNumber = types.StringValue(fmt.Sprintf("%d", fw.ServerNumber))
	data.ServerIP = types.StringValue(fw.ServerIP)
	data.Status = types.StringValue(fw.Status)
	data.AllowlistHOS = types.BoolValue(fw.AllowlistHOS)
	data.FilterIPv6 = types.BoolValue(fw.FilterIPv6)
	data.Port = types.StringValue(fw.Port)
	data.InputRules = apiRulesToModel(fw.Rules.Input)
	data.OutputRules = apiRulesToModel(fw.Rules.Output)

	if data.InputRules == nil {
		data.InputRules = []firewallRuleModel{}
	}
	if data.OutputRules == nil {
		data.OutputRules = []firewallRuleModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
