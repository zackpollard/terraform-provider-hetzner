// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &failoverResource{}
	_ resource.ResourceWithImportState = &failoverResource{}
)

func NewFailoverResource() resource.Resource {
	return &failoverResource{}
}

type failoverResource struct {
	client *client.Client
}

type failoverResourceModel struct {
	IP             types.String `tfsdk:"ip"`
	Netmask        types.String `tfsdk:"netmask"`
	ServerIP       types.String `tfsdk:"server_ip"`
	ServerIPv6     types.String `tfsdk:"server_ipv6_net"`
	ServerNumber   types.Int64  `tfsdk:"server_number"`
	ActiveServerIP types.String `tfsdk:"active_server_ip"`
}

type failoverAPIResponse struct {
	Failover failoverAPI `json:"failover"`
}

type failoverAPI struct {
	IP             string `json:"ip"`
	Netmask        string `json:"netmask"`
	ServerIP       string `json:"server_ip"`
	ServerIPv6     string `json:"server_ipv6_net"`
	ServerNumber   int    `json:"server_number"`
	ActiveServerIP string `json:"active_server_ip"`
}

func (r *failoverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failover"
}

func (r *failoverResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages failover IP routing for Hetzner dedicated servers.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The failover IP address.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active_server_ip": schema.StringAttribute{
				MarkdownDescription: "The server IP to route the failover IP to.",
				Required:            true,
			},
			"netmask": schema.StringAttribute{
				MarkdownDescription: "The failover netmask.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The owner server main IP.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "The owner server IPv6 network.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The owner server number.",
				Computed:            true,
			},
		},
	}
}

func (r *failoverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *failoverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data failoverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("active_server_ip", data.ActiveServerIP.ValueString())

	body, err := r.client.Post(fmt.Sprintf("/failover/%s", data.IP.ValueString()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error setting failover routing", err.Error())
		return
	}

	var apiResp failoverAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing failover response", err.Error())
		return
	}

	r.mapAPIToModel(&apiResp.Failover, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *failoverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data failoverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get(fmt.Sprintf("/failover/%s", data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading failover", err.Error())
		return
	}

	var apiResp failoverAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing failover response", err.Error())
		return
	}

	r.mapAPIToModel(&apiResp.Failover, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *failoverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data failoverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("active_server_ip", data.ActiveServerIP.ValueString())

	body, err := r.client.Post(fmt.Sprintf("/failover/%s", data.IP.ValueString()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating failover routing", err.Error())
		return
	}

	var apiResp failoverAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing failover response", err.Error())
		return
	}

	r.mapAPIToModel(&apiResp.Failover, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *failoverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data failoverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/failover/%s", data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting failover routing", err.Error())
		return
	}
}

func (r *failoverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}

func (r *failoverResource) mapAPIToModel(api *failoverAPI, data *failoverResourceModel) {
	data.IP = types.StringValue(api.IP)
	data.Netmask = types.StringValue(api.Netmask)
	data.ServerIP = types.StringValue(api.ServerIP)
	data.ServerIPv6 = types.StringValue(api.ServerIPv6)
	data.ServerNumber = types.Int64Value(int64(api.ServerNumber))
	data.ActiveServerIP = types.StringValue(api.ActiveServerIP)
}
