// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ ephemeral.EphemeralResource              = &wolEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &wolEphemeralResource{}
)

func NewWOLEphemeralResource() ephemeral.EphemeralResource {
	return &wolEphemeralResource{}
}

type wolEphemeralResource struct {
	client *client.Client
}

type wolEphemeralResourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
}

type wolExecuteAPIResponse struct {
	WOL wolExecuteAPIData `json:"wol"`
}

type wolExecuteAPIData struct {
	ServerIP      string `json:"server_ip"`
	ServerIPv6Net string `json:"server_ipv6_net"`
	ServerNumber  int    `json:"server_number"`
}

func (r *wolEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wol"
}

func (r *wolEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sends a Wake on LAN packet to a server. This is an ephemeral resource that triggers a one-time action and does not persist state.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number to wake.",
				Required:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Server main IPv4 address (returned after WoL).",
				Computed:            true,
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "Server IPv6 network (returned after WoL).",
				Computed:            true,
			},
		},
	}
}

func (r *wolEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	r.client = c
}

func (r *wolEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data wolEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Post(fmt.Sprintf("/wol/%d", data.ServerNumber.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error sending Wake on LAN", err.Error())
		return
	}

	var apiResp wolExecuteAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing WoL response", err.Error())
		return
	}

	data.ServerIP = types.StringValue(apiResp.WOL.ServerIP)
	data.ServerIPv6Net = types.StringValue(apiResp.WOL.ServerIPv6Net)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
