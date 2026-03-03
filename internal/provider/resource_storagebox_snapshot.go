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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &storageboxSnapshotResource{}
	_ resource.ResourceWithConfigure   = &storageboxSnapshotResource{}
	_ resource.ResourceWithImportState = &storageboxSnapshotResource{}
)

type storageboxSnapshotResource struct {
	client *client.Client
}

type storageboxSnapshotResourceModel struct {
	StorageboxID types.Int64  `tfsdk:"storagebox_id"`
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	Timestamp    types.String `tfsdk:"timestamp"`
	Size         types.Int64  `tfsdk:"size"`
}

type snapshotAPIData struct {
	Name      string  `json:"name"`
	Timestamp string  `json:"timestamp"`
	Comment   *string `json:"comment"`
	Size      int     `json:"size"`
}

type snapshotCreateAPIResponse struct {
	Snapshot snapshotAPIData `json:"snapshot"`
}

func NewStorageboxSnapshotResource() resource.Resource {
	return &storageboxSnapshotResource{}
}

func (r *storageboxSnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox_snapshot"
}

func (r *storageboxSnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages snapshots for a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Snapshot name/identifier (assigned by the API).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Snapshot comment.",
				Optional:            true,
			},
			"timestamp": schema.StringAttribute{
				MarkdownDescription: "Snapshot creation time.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Snapshot size in GB.",
				Computed:            true,
			},
		},
	}
}

func (r *storageboxSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageboxSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageboxSnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	data := url.Values{}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		data.Set("comment", plan.Comment.ValueString())
	}

	body, err := r.client.Post(fmt.Sprintf("/storagebox/%d/snapshot", sbID), data)
	if err != nil {
		resp.Diagnostics.AddError("Error creating snapshot", err.Error())
		return
	}

	var apiResp snapshotCreateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing snapshot response", err.Error())
		return
	}

	plan.Name = types.StringValue(apiResp.Snapshot.Name)
	plan.Timestamp = types.StringValue(apiResp.Snapshot.Timestamp)
	plan.Size = types.Int64Value(int64(apiResp.Snapshot.Size))
	if apiResp.Snapshot.Comment != nil {
		plan.Comment = types.StringValue(*apiResp.Snapshot.Comment)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageboxSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := state.StorageboxID.ValueInt64()
	snapName := state.Name.ValueString()

	body, err := r.client.Get(fmt.Sprintf("/storagebox/%d/snapshot", sbID))
	if err != nil {
		resp.Diagnostics.AddError("Error reading snapshots", err.Error())
		return
	}

	var snapshots []struct {
		Snapshot snapshotAPIData `json:"snapshot"`
	}
	if err := json.Unmarshal(body, &snapshots); err != nil {
		resp.Diagnostics.AddError("Error parsing snapshots response", err.Error())
		return
	}

	found := false
	for _, s := range snapshots {
		if s.Snapshot.Name == snapName {
			state.Timestamp = types.StringValue(s.Snapshot.Timestamp)
			state.Size = types.Int64Value(int64(s.Snapshot.Size))
			if s.Snapshot.Comment != nil {
				state.Comment = types.StringValue(*s.Snapshot.Comment)
			} else {
				state.Comment = types.StringNull()
			}
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageboxSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan storageboxSnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	snapName := plan.Name.ValueString()

	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		data := url.Values{}
		data.Set("comment", plan.Comment.ValueString())
		_, err := r.client.Post(fmt.Sprintf("/storagebox/%d/snapshot/%s/comment", sbID, snapName), data)
		if err != nil {
			resp.Diagnostics.AddError("Error updating snapshot comment", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageboxSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := state.StorageboxID.ValueInt64()
	snapName := state.Name.ValueString()

	_, err := r.client.Delete(fmt.Sprintf("/storagebox/%d/snapshot/%s", sbID, snapName))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting snapshot", err.Error())
		return
	}
}

func (r *storageboxSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: storagebox_id/snapshot_name
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: storagebox_id/snapshot_name")
		return
	}

	sbID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "storagebox_id must be numeric")
		return
	}

	var state storageboxSnapshotResourceModel
	state.StorageboxID = types.Int64Value(sbID)
	state.Name = types.StringValue(parts[1])
	state.Comment = types.StringNull()
	state.Timestamp = types.StringNull()
	state.Size = types.Int64Null()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
