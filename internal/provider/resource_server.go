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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

func NewServerResource() resource.Resource {
	return &serverResource{}
}

type serverResource struct {
	client *client.Client
}

type serverResourceModel struct {
	ServerNumber             types.Int64  `tfsdk:"server_number"`
	ServerName               types.String `tfsdk:"server_name"`
	ServerIP                 types.String `tfsdk:"server_ip"`
	ServerIPv6               types.String `tfsdk:"server_ipv6_net"`
	Product                  types.String `tfsdk:"product"`
	DC                       types.String `tfsdk:"dc"`
	Traffic                  types.String `tfsdk:"traffic"`
	Status                   types.String `tfsdk:"status"`
	Cancelled                types.Bool   `tfsdk:"cancelled"`
	PaidUntil                types.String `tfsdk:"paid_until"`
	CancellationDate         types.String `tfsdk:"cancellation_date"`
	ReserveLocation          types.Bool   `tfsdk:"reserve_location"`
	EarliestCancellationDate types.String `tfsdk:"earliest_cancellation_date"`
}

type serverDetailAPIResponse struct {
	Server serverDetailAPI `json:"server"`
}

type serverDetailAPI struct {
	ServerIP     string `json:"server_ip"`
	ServerIPv6   string `json:"server_ipv6_net"`
	ServerNumber int    `json:"server_number"`
	ServerName   string `json:"server_name"`
	Product      string `json:"product"`
	DC           string `json:"dc"`
	Traffic      string `json:"traffic"`
	Status       string `json:"status"`
	Cancelled    bool   `json:"cancelled"`
	PaidUntil    string `json:"paid_until"`
}

type serverCancellationAPIResponse struct {
	Cancellation serverCancellationAPI `json:"cancellation"`
}

type serverCancellationAPI struct {
	ServerIP                 string  `json:"server_ip"`
	ServerNumber             int     `json:"server_number"`
	EarliestCancellationDate string  `json:"earliest_cancellation_date"`
	Cancelled                bool    `json:"cancelled"`
	ReservationPossible      bool    `json:"reservation_possible"`
	Reserved                 bool    `json:"reserved"`
	CancellationDate         *string `json:"cancellation_date"`
}

func (r *serverResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hetzner dedicated server's name. Servers are provisioned externally; this resource only manages the server_name attribute.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The unique server number (ID).",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"server_name": schema.StringAttribute{
				MarkdownDescription: "The user-assigned server name.",
				Required:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The main IPv4 address of the server.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "The main IPv6 network of the server.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product": schema.StringAttribute{
				MarkdownDescription: "The product name.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dc": schema.StringAttribute{
				MarkdownDescription: "The data center.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"traffic": schema.StringAttribute{
				MarkdownDescription: "Free traffic quota.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Server status (ready or in process).",
				Computed:            true,
			},
			"cancelled": schema.BoolAttribute{
				MarkdownDescription: "Whether the server has been cancelled.",
				Computed:            true,
			},
			"paid_until": schema.StringAttribute{
				MarkdownDescription: "Date the server is paid until.",
				Computed:            true,
			},
			"cancellation_date": schema.StringAttribute{
				MarkdownDescription: "The cancellation date for the server (YYYY-MM-DD or \"now\"). Set to schedule cancellation, remove to revoke.",
				Optional:            true,
			},
			"reserve_location": schema.BoolAttribute{
				MarkdownDescription: "Whether to reserve the server location on cancellation.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"earliest_cancellation_date": schema.StringAttribute{
				MarkdownDescription: "The earliest possible cancellation date.",
				Computed:            true,
			},
		},
	}
}

func (r *serverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Servers are created externally. We adopt by setting the name.
	params := url.Values{}
	params.Set("server_name", data.ServerName.ValueString())

	_, err := r.client.Post(fmt.Sprintf("/server/%d", data.ServerNumber.ValueInt64()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error setting server name", err.Error())
		return
	}

	// Schedule cancellation if requested.
	if !data.CancellationDate.IsNull() {
		r.setCancellation(&data, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Read back the full server details.
	r.readServer(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readServer(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{}
	params.Set("server_name", plan.ServerName.ValueString())

	_, err := r.client.Post(fmt.Sprintf("/server/%d", plan.ServerNumber.ValueInt64()), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server name", err.Error())
		return
	}

	// Handle cancellation changes.
	planHasCancel := !plan.CancellationDate.IsNull()
	stateHasCancel := !state.CancellationDate.IsNull()

	switch {
	case planHasCancel && !stateHasCancel:
		// Schedule new cancellation.
		r.setCancellation(&plan, &resp.Diagnostics)
	case planHasCancel && stateHasCancel && plan.CancellationDate.ValueString() != state.CancellationDate.ValueString():
		// Change cancellation date: revoke then re-schedule.
		r.revokeCancellation(&plan, &resp.Diagnostics)
		if !resp.Diagnostics.HasError() {
			r.setCancellation(&plan, &resp.Diagnostics)
		}
	case !planHasCancel && stateHasCancel:
		// Revoke cancellation.
		r.revokeCancellation(&plan, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	r.readServer(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Schedule immediate cancellation.
	cancelParams := url.Values{}
	cancelParams.Set("cancellation_date", "now")

	_, err := r.client.Post(fmt.Sprintf("/server/%d/cancellation", data.ServerNumber.ValueInt64()), cancelParams)
	if err != nil {
		// Ignore 409 (already cancelled).
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 409 {
			return
		}
		resp.Diagnostics.AddError("Error cancelling server", err.Error())
		return
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "server_number must be a numeric value")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_number"), id)...)
}

func (r *serverResource) readServer(data *serverResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/server/%d", data.ServerNumber.ValueInt64()))
	if err != nil {
		diags.AddError("Error reading server", err.Error())
		return
	}

	var apiResp serverDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		diags.AddError("Error parsing server response", err.Error())
		return
	}

	s := apiResp.Server
	data.ServerNumber = types.Int64Value(int64(s.ServerNumber))
	data.ServerName = types.StringValue(s.ServerName)
	data.ServerIP = stringOrNull(s.ServerIP)
	data.ServerIPv6 = stringOrNull(s.ServerIPv6)
	data.Product = types.StringValue(s.Product)
	data.DC = types.StringValue(s.DC)
	data.Traffic = types.StringValue(s.Traffic)
	data.Status = types.StringValue(s.Status)
	data.Cancelled = types.BoolValue(s.Cancelled)
	data.PaidUntil = types.StringValue(s.PaidUntil)

	// Read cancellation state.
	cancelBody, err := r.client.Get(fmt.Sprintf("/server/%d/cancellation", data.ServerNumber.ValueInt64()))
	if err != nil {
		diags.AddError("Error reading server cancellation", err.Error())
		return
	}

	var cancelResp serverCancellationAPIResponse
	if err := json.Unmarshal(cancelBody, &cancelResp); err != nil {
		diags.AddError("Error parsing server cancellation response", err.Error())
		return
	}

	c := cancelResp.Cancellation
	data.EarliestCancellationDate = types.StringValue(c.EarliestCancellationDate)
	if c.CancellationDate != nil {
		data.CancellationDate = types.StringValue(*c.CancellationDate)
	} else {
		data.CancellationDate = types.StringNull()
	}
}

func (r *serverResource) setCancellation(data *serverResourceModel, diags *diag.Diagnostics) {
	params := url.Values{}
	params.Set("cancellation_date", data.CancellationDate.ValueString())
	if data.ReserveLocation.ValueBool() {
		params.Set("reserve_location", "true")
	}

	_, err := r.client.Post(fmt.Sprintf("/server/%d/cancellation", data.ServerNumber.ValueInt64()), params)
	if err != nil {
		diags.AddError("Error scheduling server cancellation", err.Error())
		return
	}
}

func (r *serverResource) revokeCancellation(data *serverResourceModel, diags *diag.Diagnostics) {
	_, err := r.client.Delete(fmt.Sprintf("/server/%d/cancellation", data.ServerNumber.ValueInt64()))
	if err != nil {
		diags.AddError("Error revoking server cancellation", err.Error())
		return
	}
}
