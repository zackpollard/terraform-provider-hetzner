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
	_ resource.Resource                = &storageboxSubaccountResource{}
	_ resource.ResourceWithConfigure   = &storageboxSubaccountResource{}
	_ resource.ResourceWithImportState = &storageboxSubaccountResource{}
)

type storageboxSubaccountResource struct {
	client *client.Client
}

type storageboxSubaccountResourceModel struct {
	StorageboxID         types.Int64  `tfsdk:"storagebox_id"`
	Username             types.String `tfsdk:"username"`
	Homedirectory        types.String `tfsdk:"homedirectory"`
	Samba                types.Bool   `tfsdk:"samba"`
	Webdav               types.Bool   `tfsdk:"webdav"`
	SSH                  types.Bool   `tfsdk:"ssh"`
	ExternalReachability types.Bool   `tfsdk:"external_reachability"`
	Readonly             types.Bool   `tfsdk:"readonly"`
	Createdir            types.Bool   `tfsdk:"createdir"`
	Comment              types.String `tfsdk:"comment"`
}

type subaccountAPIResponse struct {
	Subaccount subaccountAPIData `json:"subaccount"`
}

type subaccountAPIData struct {
	Username             string  `json:"username"`
	Homedirectory        string  `json:"homedirectory"`
	Samba                bool    `json:"samba"`
	Webdav               bool    `json:"webdav"`
	SSH                  bool    `json:"ssh"`
	ExternalReachability bool    `json:"external_reachability"`
	Readonly             bool    `json:"readonly"`
	Createdir            bool    `json:"createdir"`
	Comment              *string `json:"comment"`
}

func NewStorageboxSubaccountResource() resource.Resource {
	return &storageboxSubaccountResource{}
}

func (r *storageboxSubaccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storagebox_subaccount"
}

func (r *storageboxSubaccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages sub-accounts for a Hetzner Storage Box.",
		Attributes: map[string]schema.Attribute{
			"storagebox_id": schema.Int64Attribute{
				MarkdownDescription: "The storage box ID.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Sub-account username.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"homedirectory": schema.StringAttribute{
				MarkdownDescription: "Home directory path.",
				Required:            true,
			},
			"samba": schema.BoolAttribute{
				MarkdownDescription: "SMB/CIFS enabled.",
				Optional:            true,
				Computed:            true,
			},
			"webdav": schema.BoolAttribute{
				MarkdownDescription: "WebDAV enabled.",
				Optional:            true,
				Computed:            true,
			},
			"ssh": schema.BoolAttribute{
				MarkdownDescription: "SSH access enabled.",
				Optional:            true,
				Computed:            true,
			},
			"external_reachability": schema.BoolAttribute{
				MarkdownDescription: "Remote access enabled.",
				Optional:            true,
				Computed:            true,
			},
			"readonly": schema.BoolAttribute{
				MarkdownDescription: "Read-only restriction.",
				Optional:            true,
				Computed:            true,
			},
			"createdir": schema.BoolAttribute{
				MarkdownDescription: "Allow directory creation.",
				Optional:            true,
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment.",
				Optional:            true,
			},
		},
	}
}

func (r *storageboxSubaccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageboxSubaccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageboxSubaccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	data := url.Values{}
	data.Set("username", plan.Username.ValueString())
	data.Set("homedirectory", plan.Homedirectory.ValueString())

	if !plan.Samba.IsNull() && !plan.Samba.IsUnknown() {
		data.Set("samba", boolToString(plan.Samba.ValueBool()))
	}
	if !plan.Webdav.IsNull() && !plan.Webdav.IsUnknown() {
		data.Set("webdav", boolToString(plan.Webdav.ValueBool()))
	}
	if !plan.SSH.IsNull() && !plan.SSH.IsUnknown() {
		data.Set("ssh", boolToString(plan.SSH.ValueBool()))
	}
	if !plan.ExternalReachability.IsNull() && !plan.ExternalReachability.IsUnknown() {
		data.Set("external_reachability", boolToString(plan.ExternalReachability.ValueBool()))
	}
	if !plan.Readonly.IsNull() && !plan.Readonly.IsUnknown() {
		data.Set("readonly", boolToString(plan.Readonly.ValueBool()))
	}
	if !plan.Createdir.IsNull() && !plan.Createdir.IsUnknown() {
		data.Set("createdir", boolToString(plan.Createdir.ValueBool()))
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		data.Set("comment", plan.Comment.ValueString())
	}

	_, err := r.client.Post(fmt.Sprintf("/storagebox/%d/subaccount", sbID), data)
	if err != nil {
		resp.Diagnostics.AddError("Error creating sub-account", err.Error())
		return
	}

	// Read back the created subaccount
	r.readSubaccount(sbID, plan.Username.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxSubaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageboxSubaccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSubaccount(state.StorageboxID.ValueInt64(), state.Username.ValueString(), &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageboxSubaccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan storageboxSubaccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := plan.StorageboxID.ValueInt64()
	username := plan.Username.ValueString()

	data := url.Values{}
	data.Set("homedirectory", plan.Homedirectory.ValueString())

	if !plan.Samba.IsNull() && !plan.Samba.IsUnknown() {
		data.Set("samba", boolToString(plan.Samba.ValueBool()))
	}
	if !plan.Webdav.IsNull() && !plan.Webdav.IsUnknown() {
		data.Set("webdav", boolToString(plan.Webdav.ValueBool()))
	}
	if !plan.SSH.IsNull() && !plan.SSH.IsUnknown() {
		data.Set("ssh", boolToString(plan.SSH.ValueBool()))
	}
	if !plan.ExternalReachability.IsNull() && !plan.ExternalReachability.IsUnknown() {
		data.Set("external_reachability", boolToString(plan.ExternalReachability.ValueBool()))
	}
	if !plan.Readonly.IsNull() && !plan.Readonly.IsUnknown() {
		data.Set("readonly", boolToString(plan.Readonly.ValueBool()))
	}
	if !plan.Createdir.IsNull() && !plan.Createdir.IsUnknown() {
		data.Set("createdir", boolToString(plan.Createdir.ValueBool()))
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		data.Set("comment", plan.Comment.ValueString())
	}

	_, err := r.client.Put(fmt.Sprintf("/storagebox/%d/subaccount/%s", sbID, username), data)
	if err != nil {
		resp.Diagnostics.AddError("Error updating sub-account", err.Error())
		return
	}

	r.readSubaccount(sbID, username, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageboxSubaccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageboxSubaccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sbID := state.StorageboxID.ValueInt64()
	username := state.Username.ValueString()

	_, err := r.client.Delete(fmt.Sprintf("/storagebox/%d/subaccount/%s", sbID, username))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting sub-account", err.Error())
		return
	}
}

func (r *storageboxSubaccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: storagebox_id/username
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: storagebox_id/username")
		return
	}

	sbID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "storagebox_id must be numeric")
		return
	}

	var state storageboxSubaccountResourceModel
	state.StorageboxID = types.Int64Value(sbID)
	state.Username = types.StringValue(parts[1])

	r.readSubaccount(sbID, parts[1], &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageboxSubaccountResource) readSubaccount(sbID int64, username string, model *storageboxSubaccountResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/storagebox/%d/subaccount", sbID))
	if err != nil {
		diags.AddError("Error reading sub-accounts", err.Error())
		return
	}

	var subaccounts []subaccountAPIResponse
	if err := json.Unmarshal(body, &subaccounts); err != nil {
		diags.AddError("Error parsing sub-accounts response", err.Error())
		return
	}

	for _, sa := range subaccounts {
		if sa.Subaccount.Username == username {
			model.Homedirectory = types.StringValue(sa.Subaccount.Homedirectory)
			model.Samba = types.BoolValue(sa.Subaccount.Samba)
			model.Webdav = types.BoolValue(sa.Subaccount.Webdav)
			model.SSH = types.BoolValue(sa.Subaccount.SSH)
			model.ExternalReachability = types.BoolValue(sa.Subaccount.ExternalReachability)
			model.Readonly = types.BoolValue(sa.Subaccount.Readonly)
			model.Createdir = types.BoolValue(sa.Subaccount.Createdir)
			if sa.Subaccount.Comment != nil {
				model.Comment = types.StringValue(*sa.Subaccount.Comment)
			} else {
				model.Comment = types.StringNull()
			}
			return
		}
	}

	diags.AddError("Sub-account not found", fmt.Sprintf("Sub-account %q not found in storage box %d", username, sbID))
}
