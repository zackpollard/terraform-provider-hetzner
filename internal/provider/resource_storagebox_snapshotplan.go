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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &storageboxSnapshotplanResource{}
	_ resource.ResourceWithConfigure   = &storageboxSnapshotplanResource{}
	_ resource.ResourceWithImportState = &storageboxSnapshotplanResource{}
)

type storageboxSnapshotplanResource struct {
	client *client.Client
}

type storageboxSnapshotplanResourceModel struct {
	StorageboxID types.Int64  `tfsdk:"storagebox_id"`
	Status       types.String `tfsdk:"status"`
	Minute       types.Int64  `tfsdk:"minute"`
	Hour         types.Int64  `tfsdk:"hour"`
	DayOfWeek    types.Int64  `tfsdk:"day_of_week"`
	DayOfMonth   types.Int64  `tfsdk:"day_of_month"`
	Month        types.Int64  `tfsdk:"month"`
}

type snapshotplanAPIResponse struct {
	Snapshotplan snapshotplanAPIData `json:"snapshotplan"`
}

type snapshotplanAPIData struct {
	Status     string `json:"status"`
	Minute     int    `json:"minute"`
	Hour       int    `json:"hour"`
	DayOfWeek  int    `json:"day_of_week"`
	DayOfMonth int    `json:"day_of_month"`
	Month      int    `json:"month"`
}

func NewStorageboxSnapshotplanResource() resource.Resource {
	return &storageboxSnapshotplanResource{}
}

func (r *storageboxSnapshotplanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox_snapshotplan"
}

func (r *storageboxSnapshotplanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the snapshot plan (automated snapshot schedule) for a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Plan status: `enabled` or `disabled`.",
				Required:            true,
			},
			"minute": schema.Int64Attribute{
				MarkdownDescription: "Minute (0-59).",
				Optional:            true,
				Computed:            true,
			},
			"hour": schema.Int64Attribute{
				MarkdownDescription: "Hour (0-23).",
				Optional:            true,
				Computed:            true,
			},
			"day_of_week": schema.Int64Attribute{
				MarkdownDescription: "Day of week (0=Sun, 6=Sat).",
				Optional:            true,
				Computed:            true,
			},
			"day_of_month": schema.Int64Attribute{
				MarkdownDescription: "Day of month (1-31).",
				Optional:            true,
				Computed:            true,
			},
			"month": schema.Int64Attribute{
				MarkdownDescription: "Month (1-12).",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *storageboxSnapshotplanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageboxSnapshotplanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageboxSnapshotplanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	data := r.buildFormData(&plan)

	_, err := r.client.Post(fmt.Sprintf("/storagebox/%d/snapshotplan", sbID), data)
	if err != nil {
		resp.Diagnostics.AddError("Error setting snapshot plan", err.Error())
		return
	}

	r.readInto(sbID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxSnapshotplanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageboxSnapshotplanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readInto(state.StorageboxID.ValueInt64(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageboxSnapshotplanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan storageboxSnapshotplanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	data := r.buildFormData(&plan)

	_, err := r.client.Post(fmt.Sprintf("/storagebox/%d/snapshotplan", sbID), data)
	if err != nil {
		resp.Diagnostics.AddError("Error updating snapshot plan", err.Error())
		return
	}

	r.readInto(sbID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxSnapshotplanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageboxSnapshotplanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data := url.Values{}
	data.Set("status", "disabled")

	_, err := r.client.Post(fmt.Sprintf("/storagebox/%d/snapshotplan", state.StorageboxID.ValueInt64()), data)
	if err != nil {
		resp.Diagnostics.AddError("Error disabling snapshot plan", err.Error())
		return
	}
}

func (r *storageboxSnapshotplanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sbID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected numeric storagebox_id")
		return
	}

	var state storageboxSnapshotplanResourceModel
	state.StorageboxID = types.Int64Value(sbID)

	r.readInto(sbID, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageboxSnapshotplanResource) buildFormData(plan *storageboxSnapshotplanResourceModel) url.Values {
	data := url.Values{}
	data.Set("status", plan.Status.ValueString())

	if !plan.Minute.IsNull() && !plan.Minute.IsUnknown() {
		data.Set("minute", strconv.FormatInt(plan.Minute.ValueInt64(), 10))
	}
	if !plan.Hour.IsNull() && !plan.Hour.IsUnknown() {
		data.Set("hour", strconv.FormatInt(plan.Hour.ValueInt64(), 10))
	}
	if !plan.DayOfWeek.IsNull() && !plan.DayOfWeek.IsUnknown() {
		data.Set("day_of_week", strconv.FormatInt(plan.DayOfWeek.ValueInt64(), 10))
	}
	if !plan.DayOfMonth.IsNull() && !plan.DayOfMonth.IsUnknown() {
		data.Set("day_of_month", strconv.FormatInt(plan.DayOfMonth.ValueInt64(), 10))
	}
	if !plan.Month.IsNull() && !plan.Month.IsUnknown() {
		data.Set("month", strconv.FormatInt(plan.Month.ValueInt64(), 10))
	}

	return data
}

func (r *storageboxSnapshotplanResource) readInto(sbID int64, model *storageboxSnapshotplanResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/storagebox/%d/snapshotplan", sbID))
	if err != nil {
		diags.AddError("Error reading snapshot plan", err.Error())
		return
	}

	var apiResp snapshotplanAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		diags.AddError("Error parsing snapshot plan response", err.Error())
		return
	}

	sp := apiResp.Snapshotplan
	model.Status = types.StringValue(sp.Status)
	model.Minute = types.Int64Value(int64(sp.Minute))
	model.Hour = types.Int64Value(int64(sp.Hour))
	model.DayOfWeek = types.Int64Value(int64(sp.DayOfWeek))
	model.DayOfMonth = types.Int64Value(int64(sp.DayOfMonth))
	model.Month = types.Int64Value(int64(sp.Month))
}
