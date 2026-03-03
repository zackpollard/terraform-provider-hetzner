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
	_ resource.Resource                = &ipResource{}
	_ resource.ResourceWithImportState = &ipResource{}
)

func NewIPResource() resource.Resource {
	return &ipResource{}
}

type ipResource struct {
	client *client.Client
}

type ipResourceModel struct {
	IP              types.String `tfsdk:"ip"`
	ServerIP        types.String `tfsdk:"server_ip"`
	ServerNumber    types.Int64  `tfsdk:"server_number"`
	Locked          types.Bool   `tfsdk:"locked"`
	SeparateMAC     types.String `tfsdk:"separate_mac"`
	TrafficWarnings types.Bool   `tfsdk:"traffic_warnings"`
	TrafficHourly   types.Int64  `tfsdk:"traffic_hourly"`
	TrafficDaily    types.Int64  `tfsdk:"traffic_daily"`
	TrafficMonthly  types.Int64  `tfsdk:"traffic_monthly"`
	Gateway         types.String `tfsdk:"gateway"`
	Mask            types.Int64  `tfsdk:"mask"`
	Broadcast       types.String `tfsdk:"broadcast"`
}

func (r *ipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip"
}

func (r *ipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages traffic warning configuration for a Hetzner IP address. IPs are provisioned externally; this resource manages traffic warning settings.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address.",
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
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The server main IP this IP is assigned to.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server ID this IP is assigned to.",
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether the IP is locked.",
				Computed:            true,
			},
			"separate_mac": schema.StringAttribute{
				MarkdownDescription: "Separate MAC address, if assigned.",
				Computed:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mask": schema.Int64Attribute{
				MarkdownDescription: "CIDR notation.",
				Computed:            true,
			},
			"broadcast": schema.StringAttribute{
				MarkdownDescription: "Broadcast address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// IPs are provisioned externally. We adopt by updating traffic warning settings.
	r.updateTrafficWarnings(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIP(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIP(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateTrafficWarnings(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIP(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// IPs cannot be deleted via API. We just remove from state.
	// Traffic warnings revert to API defaults naturally.
}

func (r *ipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}

func (r *ipResource) updateTrafficWarnings(data *ipResourceModel, diags *diag.Diagnostics) {
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

	_, err := r.client.Post(fmt.Sprintf("/ip/%s", data.IP.ValueString()), params)
	if err != nil {
		diags.AddError("Error updating IP traffic warnings", err.Error())
	}
}

func (r *ipResource) readIP(data *ipResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/ip/%s", data.IP.ValueString()))
	if err != nil {
		diags.AddError("Error reading IP", err.Error())
		return
	}

	var apiResp ipDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		diags.AddError("Error parsing IP response", err.Error())
		return
	}

	ip := apiResp.IP
	data.IP = types.StringValue(ip.IP)
	data.ServerIP = types.StringValue(ip.ServerIP)
	data.ServerNumber = types.Int64Value(int64(ip.ServerNumber))
	data.Locked = types.BoolValue(ip.Locked)
	if ip.SeparateMAC != nil {
		data.SeparateMAC = types.StringValue(*ip.SeparateMAC)
	} else {
		data.SeparateMAC = types.StringNull()
	}
	data.TrafficWarnings = types.BoolValue(ip.TrafficWarnings)
	data.TrafficHourly = types.Int64Value(int64(ip.TrafficHourly))
	data.TrafficDaily = types.Int64Value(int64(ip.TrafficDaily))
	data.TrafficMonthly = types.Int64Value(int64(ip.TrafficMonthly))
	data.Gateway = types.StringValue(ip.Gateway)
	data.Mask = types.Int64Value(int64(ip.Mask))
	data.Broadcast = types.StringValue(ip.Broadcast)
}
