// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &vSwitchResource{}
	_ resource.ResourceWithImportState = &vSwitchResource{}
)

func NewVSwitchResource() resource.Resource {
	return &vSwitchResource{}
}

type vSwitchResource struct {
	client *client.Client
}

type vSwitchResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Vlan      types.Int64  `tfsdk:"vlan"`
	Cancelled types.Bool   `tfsdk:"cancelled"`
}

type vSwitchAPIResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Vlan      int    `json:"vlan"`
	Cancelled bool   `json:"cancelled"`
}

func (r *vSwitchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vswitch"
}

func (r *vSwitchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hetzner vSwitch for Layer 2 networking between dedicated servers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The unique ID of the vSwitch.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the vSwitch.",
				Required:            true,
			},
			"vlan": schema.Int64Attribute{
				MarkdownDescription: "The VLAN ID (4000-4091).",
				Required:            true,
			},
			"cancelled": schema.BoolAttribute{
				MarkdownDescription: "Whether the vSwitch has been cancelled.",
				Computed:            true,
			},
		},
	}
}

func (r *vSwitchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vSwitchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vSwitchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("name", data.Name.ValueString())
	params.Set("vlan", strconv.FormatInt(data.Vlan.ValueInt64(), 10))

	body, err := r.client.Post("/vswitch", params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating vSwitch", err.Error())
		return
	}

	var apiResp vSwitchAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing vSwitch response", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(apiResp.ID))
	data.Name = types.StringValue(apiResp.Name)
	data.Vlan = types.Int64Value(int64(apiResp.Vlan))
	data.Cancelled = types.BoolValue(apiResp.Cancelled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vSwitchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vSwitchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get(fmt.Sprintf("/vswitch/%d", data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading vSwitch", err.Error())
		return
	}

	var apiResp vSwitchAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing vSwitch response", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(apiResp.ID))
	data.Name = types.StringValue(apiResp.Name)
	data.Vlan = types.Int64Value(int64(apiResp.Vlan))
	data.Cancelled = types.BoolValue(apiResp.Cancelled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vSwitchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data vSwitchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("name", data.Name.ValueString())
	params.Set("vlan", strconv.FormatInt(data.Vlan.ValueInt64(), 10))

	_, err := r.client.Post(fmt.Sprintf("/vswitch/%d", data.ID.ValueInt64()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating vSwitch", err.Error())
		return
	}

	// Re-read to get the canonical state.
	body, err := r.client.Get(fmt.Sprintf("/vswitch/%d", data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading vSwitch after update", err.Error())
		return
	}

	var apiResp vSwitchAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing vSwitch response", err.Error())
		return
	}

	data.ID = types.Int64Value(int64(apiResp.ID))
	data.Name = types.StringValue(apiResp.Name)
	data.Vlan = types.Int64Value(int64(apiResp.Vlan))
	data.Cancelled = types.BoolValue(apiResp.Cancelled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vSwitchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vSwitchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("cancellation_date", "now")

	_, err := r.client.DeleteWithBody(fmt.Sprintf("/vswitch/%d", data.ID.ValueInt64()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting vSwitch", err.Error())
		return
	}
}

func (r *vSwitchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "vSwitch ID must be a numeric value")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
