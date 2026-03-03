// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &subnetResource{}
	_ resource.ResourceWithImportState = &subnetResource{}
)

func NewSubnetResource() resource.Resource {
	return &subnetResource{}
}

type subnetResource struct {
	client *client.Client
}

type subnetResourceModel struct {
	IP              types.String `tfsdk:"ip"`
	Mask            types.Int64  `tfsdk:"mask"`
	Gateway         types.String `tfsdk:"gateway"`
	ServerIP        types.String `tfsdk:"server_ip"`
	ServerNumber    types.Int64  `tfsdk:"server_number"`
	Failover        types.Bool   `tfsdk:"failover"`
	Locked          types.Bool   `tfsdk:"locked"`
	TrafficWarnings types.Bool   `tfsdk:"traffic_warnings"`
	TrafficHourly   types.Int64  `tfsdk:"traffic_hourly"`
	TrafficDaily    types.Int64  `tfsdk:"traffic_daily"`
	TrafficMonthly  types.Int64  `tfsdk:"traffic_monthly"`
}

func (r *subnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages traffic warning configuration for a Hetzner subnet. Subnets are provisioned externally; this resource manages traffic warning settings.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The subnet IP address.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"traffic_warnings": schema.BoolAttribute{
				MarkdownDescription: "Whether traffic warnings are enabled.",
				Required:            true,
			},
			"traffic_hourly": schema.Int64Attribute{
				MarkdownDescription: "Hourly traffic limit in MB.",
				Required:            true,
			},
			"traffic_daily": schema.Int64Attribute{
				MarkdownDescription: "Daily traffic limit in MB.",
				Required:            true,
			},
			"traffic_monthly": schema.Int64Attribute{
				MarkdownDescription: "Monthly traffic limit in GB.",
				Required:            true,
			},
			"mask": schema.Int64Attribute{
				MarkdownDescription: "The CIDR notation.",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The subnet gateway.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The server main IP this subnet is assigned to.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server ID this subnet is assigned to.",
				Computed:            true,
			},
			"failover": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a failover subnet.",
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether the subnet is locked.",
				Computed:            true,
			},
		},
	}
}

func (r *subnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Subnets are provisioned externally. We adopt by updating traffic warning settings.
	r.updateTrafficWarnings(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSubnet(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSubnet(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data subnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateTrafficWarnings(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSubnet(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Subnets cannot be deleted via API. We just remove from state.
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}

func (r *subnetResource) updateTrafficWarnings(data *subnetResourceModel, diags *diag.Diagnostics) {
	params := url.Values{}
	if !data.TrafficWarnings.IsNull() {
		params.Set("traffic_warnings", strconv.FormatBool(data.TrafficWarnings.ValueBool()))
	}
	if !data.TrafficHourly.IsNull() {
		params.Set("traffic_hourly", strconv.FormatInt(data.TrafficHourly.ValueInt64(), 10))
	}
	if !data.TrafficDaily.IsNull() {
		params.Set("traffic_daily", strconv.FormatInt(data.TrafficDaily.ValueInt64(), 10))
	}
	if !data.TrafficMonthly.IsNull() {
		params.Set("traffic_monthly", strconv.FormatInt(data.TrafficMonthly.ValueInt64(), 10))
	}

	_, err := r.client.Post(fmt.Sprintf("/subnet/%s", data.IP.ValueString()), params)
	if err != nil {
		diags.AddError("Error updating subnet traffic warnings", err.Error())
	}
}

func (r *subnetResource) readSubnet(data *subnetResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/subnet/%s", data.IP.ValueString()))
	if err != nil {
		diags.AddError("Error reading subnet", err.Error())
		return
	}

	var apiResp subnetDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		diags.AddError("Error parsing subnet response", err.Error())
		return
	}

	s := apiResp.Subnet
	data.IP = types.StringValue(s.IP)
	data.Mask = types.Int64Value(int64(s.Mask))
	data.Gateway = types.StringValue(s.Gateway)
	data.ServerIP = types.StringValue(s.ServerIP)
	data.ServerNumber = types.Int64Value(int64(s.ServerNumber))
	data.Failover = types.BoolValue(s.Failover)
	data.Locked = types.BoolValue(s.Locked)
	data.TrafficWarnings = types.BoolValue(s.TrafficWarnings)
	data.TrafficHourly = types.Int64Value(int64(s.TrafficHourly))
	data.TrafficDaily = types.Int64Value(int64(s.TrafficDaily))
	data.TrafficMonthly = types.Int64Value(int64(s.TrafficMonthly))
}
