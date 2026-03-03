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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &storageboxResource{}
	_ resource.ResourceWithConfigure   = &storageboxResource{}
	_ resource.ResourceWithImportState = &storageboxResource{}
)

type storageboxResource struct {
	client *client.Client
}

type storageboxResourceModel struct {
	StorageboxID         types.Int64  `tfsdk:"storagebox_id"`
	StorageboxName       types.String `tfsdk:"storagebox_name"`
	Webdav               types.Bool   `tfsdk:"webdav"`
	Samba                types.Bool   `tfsdk:"samba"`
	SSH                  types.Bool   `tfsdk:"ssh"`
	ExternalReachability types.Bool   `tfsdk:"external_reachability"`
	ZFS                  types.Bool   `tfsdk:"zfs"`
	DiskQuota            types.Int64  `tfsdk:"disk_quota"`
	DiskUsage            types.Int64  `tfsdk:"disk_usage"`
	Status               types.String `tfsdk:"status"`
	PaidUntil            types.String `tfsdk:"paid_until"`
	Locked               types.Bool   `tfsdk:"locked"`
	Server               types.Int64  `tfsdk:"server"`
}

type storageboxAPIResponse struct {
	Storagebox storageboxAPIData `json:"storagebox"`
}

type storageboxAPIData struct {
	StorageboxID         int    `json:"storagebox_id"`
	StorageboxName       string `json:"storagebox_name"`
	DiskQuota            int    `json:"disk_quota"`
	DiskUsage            int    `json:"disk_usage"`
	Status               string `json:"status"`
	PaidUntil            string `json:"paid_until"`
	Locked               bool   `json:"locked"`
	Server               *int   `json:"server"`
	Webdav               bool   `json:"webdav"`
	Samba                bool   `json:"samba"`
	SSH                  bool   `json:"ssh"`
	ExternalReachability bool   `json:"external_reachability"`
	ZFS                  bool   `json:"zfs"`
}

func NewStorageboxResource() resource.Resource {
	return &storageboxResource{}
}

func (r *storageboxResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox"
}

func (r *storageboxResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages settings for a Hetzner Storage Box. Storage boxes are provisioned externally; this resource manages configuration only.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"storagebox_name": schema.StringAttribute{
				MarkdownDescription: "Name of the storage box.",
				Optional:            true,
				Computed:            true,
			},
			"webdav": schema.BoolAttribute{
				MarkdownDescription: "Whether WebDAV is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"samba": schema.BoolAttribute{
				MarkdownDescription: "Whether SMB/CIFS is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"ssh": schema.BoolAttribute{
				MarkdownDescription: "Whether SSH access is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"external_reachability": schema.BoolAttribute{
				MarkdownDescription: "Whether remote access is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"zfs": schema.BoolAttribute{
				MarkdownDescription: "Whether ZFS features are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"disk_quota": schema.Int64Attribute{
				MarkdownDescription: "Total disk capacity in GB.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"disk_usage": schema.Int64Attribute{
				MarkdownDescription: "Used disk space in GB.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Operational status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"paid_until": schema.StringAttribute{
				MarkdownDescription: "Expiration date (yyyy-MM-dd).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether access is restricted.",
				Computed:            true,
			},
			"server": schema.Int64Attribute{
				MarkdownDescription: "Linked server ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *storageboxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageboxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageboxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()

	// Update settings via POST
	data := url.Values{}
	if !plan.StorageboxName.IsNull() && !plan.StorageboxName.IsUnknown() {
		data.Set("storagebox_name", plan.StorageboxName.ValueString())
	}
	if !plan.Webdav.IsNull() && !plan.Webdav.IsUnknown() {
		data.Set("webdav", boolToString(plan.Webdav.ValueBool()))
	}
	if !plan.Samba.IsNull() && !plan.Samba.IsUnknown() {
		data.Set("samba", boolToString(plan.Samba.ValueBool()))
	}
	if !plan.SSH.IsNull() && !plan.SSH.IsUnknown() {
		data.Set("ssh", boolToString(plan.SSH.ValueBool()))
	}
	if !plan.ExternalReachability.IsNull() && !plan.ExternalReachability.IsUnknown() {
		data.Set("external_reachability", boolToString(plan.ExternalReachability.ValueBool()))
	}
	if !plan.ZFS.IsNull() && !plan.ZFS.IsUnknown() {
		data.Set("zfs", boolToString(plan.ZFS.ValueBool()))
	}

	if len(data) > 0 {
		_, err := r.client.Post(fmt.Sprintf("/storagebox/%d", sbID), data)
		if err != nil {
			resp.Diagnostics.AddError("Error updating storage box settings", err.Error())
			return
		}
	}

	// Read back the current state
	r.readInto(sbID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageboxResourceModel
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

func (r *storageboxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan storageboxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	data := url.Values{}

	if !plan.StorageboxName.IsNull() && !plan.StorageboxName.IsUnknown() {
		data.Set("storagebox_name", plan.StorageboxName.ValueString())
	}
	if !plan.Webdav.IsNull() && !plan.Webdav.IsUnknown() {
		data.Set("webdav", boolToString(plan.Webdav.ValueBool()))
	}
	if !plan.Samba.IsNull() && !plan.Samba.IsUnknown() {
		data.Set("samba", boolToString(plan.Samba.ValueBool()))
	}
	if !plan.SSH.IsNull() && !plan.SSH.IsUnknown() {
		data.Set("ssh", boolToString(plan.SSH.ValueBool()))
	}
	if !plan.ExternalReachability.IsNull() && !plan.ExternalReachability.IsUnknown() {
		data.Set("external_reachability", boolToString(plan.ExternalReachability.ValueBool()))
	}
	if !plan.ZFS.IsNull() && !plan.ZFS.IsUnknown() {
		data.Set("zfs", boolToString(plan.ZFS.ValueBool()))
	}

	if len(data) > 0 {
		_, err := r.client.Post(fmt.Sprintf("/storagebox/%d", sbID), data)
		if err != nil {
			resp.Diagnostics.AddError("Error updating storage box settings", err.Error())
			return
		}
	}

	r.readInto(sbID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Storage boxes are externally provisioned. Delete simply removes from state.
}

func (r *storageboxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sbID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected numeric storagebox_id")
		return
	}

	var state storageboxResourceModel
	state.StorageboxID = types.Int64Value(sbID)

	r.readInto(sbID, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageboxResource) readInto(sbID int64, model *storageboxResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/storagebox/%d", sbID))
	if err != nil {
		diags.AddError("Error reading storage box", err.Error())
		return
	}

	var apiResp storageboxAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		diags.AddError("Error parsing storage box response", err.Error())
		return
	}

	sb := apiResp.Storagebox
	model.StorageboxID = types.Int64Value(int64(sb.StorageboxID))
	model.StorageboxName = types.StringValue(sb.StorageboxName)
	model.Webdav = types.BoolValue(sb.Webdav)
	model.Samba = types.BoolValue(sb.Samba)
	model.SSH = types.BoolValue(sb.SSH)
	model.ExternalReachability = types.BoolValue(sb.ExternalReachability)
	model.ZFS = types.BoolValue(sb.ZFS)
	model.DiskQuota = types.Int64Value(int64(sb.DiskQuota))
	model.DiskUsage = types.Int64Value(int64(sb.DiskUsage))
	model.Status = types.StringValue(sb.Status)
	model.PaidUntil = types.StringValue(sb.PaidUntil)
	model.Locked = types.BoolValue(sb.Locked)

	if sb.Server != nil {
		model.Server = types.Int64Value(int64(*sb.Server))
	} else {
		model.Server = types.Int64Null()
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
