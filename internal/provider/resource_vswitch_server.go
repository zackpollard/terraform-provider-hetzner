// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &vSwitchServerResource{}
	_ resource.ResourceWithImportState = &vSwitchServerResource{}
)

func NewVSwitchServerResource() resource.Resource {
	return &vSwitchServerResource{}
}

type vSwitchServerResource struct {
	client *client.Client
}

type vSwitchServerResourceModel struct {
	VSwitchID    types.Int64  `tfsdk:"vswitch_id"`
	ServerNumber types.Int64  `tfsdk:"server_number"`
	Status       types.String `tfsdk:"status"`
}

// vSwitchDetailAPIResponse is the full response from GET /vswitch/{id}.
type vSwitchDetailAPIResponse struct {
	ID        int                  `json:"id"`
	Name      string               `json:"name"`
	Vlan      int                  `json:"vlan"`
	Cancelled bool                 `json:"cancelled"`
	Server    []vSwitchServerEntry `json:"server"`
}

type vSwitchServerEntry struct {
	ServerNumber int    `json:"server_number"`
	ServerIP     string `json:"server_ip"`
	ServerIPv6   string `json:"server_ipv6_net"`
	Status       string `json:"status"`
}

func (r *vSwitchServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vswitch_server"
}

func (r *vSwitchServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the association of a server to a Hetzner vSwitch.",
		Attributes: map[string]schema.Attribute{
			"vswitch_id": schema.Int64Attribute{
				MarkdownDescription: "The vSwitch ID.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number to connect to the vSwitch.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The connection status (ready, in process, or failed).",
				Computed:            true,
			},
		},
	}
}

func (r *vSwitchServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vSwitchServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vSwitchServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("server", strconv.FormatInt(data.ServerNumber.ValueInt64(), 10))

	_, err := r.client.Post(fmt.Sprintf("/vswitch/%d/server", data.VSwitchID.ValueInt64()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error adding server to vSwitch", err.Error())
		return
	}

	// Read the vSwitch to get the server's connection status.
	status, err := r.readServerStatus(data.VSwitchID.ValueInt64(), data.ServerNumber.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading vSwitch server status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vSwitchServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vSwitchServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.readServerStatus(data.VSwitchID.ValueInt64(), data.ServerNumber.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error reading vSwitch server status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vSwitchServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Both attributes require replace, so Update should never be called.
	resp.Diagnostics.AddError("Unexpected update", "vswitch_server resource does not support in-place updates")
}

func (r *vSwitchServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vSwitchServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("server", strconv.FormatInt(data.ServerNumber.ValueInt64(), 10))

	_, err := r.client.DeleteWithBody(fmt.Sprintf("/vswitch/%d/server", data.VSwitchID.ValueInt64()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error removing server from vSwitch", err.Error())
		return
	}
}

func (r *vSwitchServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "vswitch_id/server_number"
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: vswitch_id/server_number")
		return
	}

	vswitchID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "vswitch_id must be numeric")
		return
	}

	serverNumber, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "server_number must be numeric")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vswitch_id"), vswitchID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_number"), serverNumber)...)
}

func (r *vSwitchServerResource) readServerStatus(vswitchID, serverNumber int64) (string, error) {
	body, err := r.client.Get(fmt.Sprintf("/vswitch/%d", vswitchID))
	if err != nil {
		return "", err
	}

	var apiResp vSwitchDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("parsing vSwitch response: %w", err)
	}

	for _, s := range apiResp.Server {
		if int64(s.ServerNumber) == serverNumber {
			return s.Status, nil
		}
	}

	return "", fmt.Errorf("server %d not found in vSwitch %d", serverNumber, vswitchID)
}
