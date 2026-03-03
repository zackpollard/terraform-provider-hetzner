// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &sshKeyResource{}
	_ resource.ResourceWithImportState = &sshKeyResource{}
	_ resource.ResourceWithConfigure   = &sshKeyResource{}
)

type sshKeyResource struct {
	client *client.Client
}

type sshKeyResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Type        types.String `tfsdk:"type"`
	Size        types.Int64  `tfsdk:"size"`
	Data        types.String `tfsdk:"data"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

type sshKeyAPIResponse struct {
	Key sshKeyAPIModel `json:"key"`
}

type sshKeyAPIModel struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	Data        string `json:"data"`
	CreatedAt   string `json:"created_at"`
}

func NewSSHKeyResource() resource.Resource {
	return &sshKeyResource{}
}

func (r *sshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *sshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an SSH public key stored in the Hetzner Robot account.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the SSH key.",
				Required:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "Fingerprint of the SSH key. Used as the unique identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Algorithm type: RSA, ECDSA, or ED25519.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits.",
				Computed:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Public key in OpenSSH format. Immutable after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation date of the SSH key.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *sshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sshKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("name", data.Name.ValueString())
	form.Set("data", data.Data.ValueString())

	body, err := r.client.Post("/key", form)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SSH key", err.Error())
		return
	}

	var apiResp sshKeyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing SSH key response", err.Error())
		return
	}

	data.Fingerprint = types.StringValue(apiResp.Key.Fingerprint)
	data.Type = types.StringValue(apiResp.Key.Type)
	data.Size = types.Int64Value(apiResp.Key.Size)
	data.Data = types.StringValue(apiResp.Key.Data)
	data.Name = types.StringValue(apiResp.Key.Name)
	data.CreatedAt = types.StringValue(apiResp.Key.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data sshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get("/key/" + url.PathEscape(data.Fingerprint.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SSH key", err.Error())
		return
	}

	var apiResp sshKeyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing SSH key response", err.Error())
		return
	}

	data.Name = types.StringValue(apiResp.Key.Name)
	data.Fingerprint = types.StringValue(apiResp.Key.Fingerprint)
	data.Type = types.StringValue(apiResp.Key.Type)
	data.Size = types.Int64Value(apiResp.Key.Size)
	data.Data = types.StringValue(apiResp.Key.Data)
	data.CreatedAt = types.StringValue(apiResp.Key.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data sshKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state sshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("name", data.Name.ValueString())

	body, err := r.client.Post("/key/"+url.PathEscape(state.Fingerprint.ValueString()), form)
	if err != nil {
		resp.Diagnostics.AddError("Error updating SSH key", err.Error())
		return
	}

	var apiResp sshKeyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing SSH key response", err.Error())
		return
	}

	data.Fingerprint = types.StringValue(apiResp.Key.Fingerprint)
	data.Type = types.StringValue(apiResp.Key.Type)
	data.Size = types.Int64Value(apiResp.Key.Size)
	data.Data = types.StringValue(apiResp.Key.Data)
	data.Name = types.StringValue(apiResp.Key.Name)
	data.CreatedAt = types.StringValue(apiResp.Key.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete("/key/" + url.PathEscape(data.Fingerprint.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error deleting SSH key", err.Error())
	}
}

func (r *sshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, sshKeyFingerprintPath, req, resp)
}

var sshKeyFingerprintPath = frameworkPath("fingerprint")
