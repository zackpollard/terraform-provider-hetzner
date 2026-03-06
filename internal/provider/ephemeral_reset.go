// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ ephemeral.EphemeralResource              = &resetEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &resetEphemeralResource{}
)

func NewResetEphemeralResource() ephemeral.EphemeralResource {
	return &resetEphemeralResource{}
}

type resetEphemeralResource struct {
	client *client.Client
}

type resetEphemeralResourceModel struct {
	ServerNumber types.Int64  `tfsdk:"server_number"`
	Type         types.String `tfsdk:"type"`
	ServerIP     types.String `tfsdk:"server_ip"`
	ServerIPv6   types.String `tfsdk:"server_ipv6_net"`
}

type resetExecuteAPIResponse struct {
	Reset resetExecuteAPIData `json:"reset"`
}

type resetExecuteAPIData struct {
	ServerIP      string `json:"server_ip"`
	ServerIPv6Net string `json:"server_ipv6_net"`
	ServerNumber  int    `json:"server_number"`
	Type          string `json:"type"`
}

func (r *resetEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reset"
}

func (r *resetEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Executes a server reset. This is an ephemeral resource that triggers a one-time action and does not persist state.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number to reset.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Reset type: \"sw\" (software), \"hw\" (hardware), or \"man\" (manual).",
				Required:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Server main IPv4 address (returned after reset).",
				Computed:            true,
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "Server IPv6 network (returned after reset).",
				Computed:            true,
			},
		},
	}
}

func (r *resetEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *resetEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data resetEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("type", data.Type.ValueString())

	body, err := r.client.Post(fmt.Sprintf("/reset/%d", data.ServerNumber.ValueInt64()), form)
	if err != nil {
		resp.Diagnostics.AddError("Error executing reset", err.Error())
		return
	}

	var apiResp resetExecuteAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing reset response", err.Error())
		return
	}

	data.ServerIP = types.StringValue(apiResp.Reset.ServerIP)
	data.ServerIPv6 = types.StringValue(apiResp.Reset.ServerIPv6Net)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
