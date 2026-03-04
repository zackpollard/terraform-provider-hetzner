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
	_ resource.Resource                = &rdnsResource{}
	_ resource.ResourceWithImportState = &rdnsResource{}
	_ resource.ResourceWithConfigure   = &rdnsResource{}
)

type rdnsResource struct {
	client *client.Client
}

type rdnsResourceModel struct {
	IP  types.String `tfsdk:"ip"`
	PTR types.String `tfsdk:"ptr"`
}

type rdnsAPIResponse struct {
	Rdns rdnsAPIModel `json:"rdns"`
}

type rdnsAPIModel struct {
	IP  string `json:"ip"`
	PTR string `json:"ptr"`
}

func NewRDNSResource() resource.Resource {
	return &rdnsResource{}
}

func (r *rdnsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdns"
}

func (r *rdnsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a reverse DNS (PTR) record for an IP address.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "IP address for the rDNS entry. Used as the unique identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ptr": schema.StringAttribute{
				MarkdownDescription: "PTR record value (hostname).",
				Required:            true,
			},
		},
	}
}

func (r *rdnsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *rdnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data rdnsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("ptr", data.PTR.ValueString())

	// Use POST (not PUT) because servers may already have a default rDNS entry.
	// PUT returns 409 RDNS_ALREADY_EXISTS, while POST creates or updates.
	body, err := r.client.Post("/rdns/"+url.PathEscape(data.IP.ValueString()), form)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rDNS entry", err.Error())
		return
	}

	var apiResp rdnsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing rDNS response", err.Error())
		return
	}

	data.IP = types.StringValue(apiResp.Rdns.IP)
	data.PTR = types.StringValue(apiResp.Rdns.PTR)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *rdnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data rdnsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get("/rdns/" + url.PathEscape(data.IP.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading rDNS entry", err.Error())
		return
	}

	var apiResp rdnsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing rDNS response", err.Error())
		return
	}

	data.IP = types.StringValue(apiResp.Rdns.IP)
	data.PTR = types.StringValue(apiResp.Rdns.PTR)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *rdnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data rdnsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("ptr", data.PTR.ValueString())

	body, err := r.client.Post("/rdns/"+url.PathEscape(data.IP.ValueString()), form)
	if err != nil {
		resp.Diagnostics.AddError("Error updating rDNS entry", err.Error())
		return
	}

	var apiResp rdnsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing rDNS response", err.Error())
		return
	}

	data.IP = types.StringValue(apiResp.Rdns.IP)
	data.PTR = types.StringValue(apiResp.Rdns.PTR)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *rdnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data rdnsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete("/rdns/" + url.PathEscape(data.IP.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error deleting rDNS entry", err.Error())
	}
}

func (r *rdnsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath("ip"), req, resp)
}
