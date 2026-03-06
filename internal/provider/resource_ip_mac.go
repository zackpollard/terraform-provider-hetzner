// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &ipMACResource{}
	_ resource.ResourceWithImportState = &ipMACResource{}
)

func NewIPMACResource() resource.Resource {
	return &ipMACResource{}
}

type ipMACResource struct {
	client *client.Client
}

type ipMACResourceModel struct {
	IP  types.String `tfsdk:"ip"`
	MAC types.String `tfsdk:"mac"`
}

type ipMACAPIResponse struct {
	MAC ipMACAPIData `json:"mac"`
}

type ipMACAPIData struct {
	IP  string `json:"ip"`
	MAC string `json:"mac"`
}

func (r *ipMACResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_mac"
}

func (r *ipMACResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a separate MAC address for a Hetzner IP address.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address to generate a separate MAC for.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac": schema.StringAttribute{
				MarkdownDescription: "The generated MAC address.",
				Computed:            true,
			},
		},
	}
}

func (r *ipMACResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ipMACResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ipMACResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Put(fmt.Sprintf("/ip/%s/mac", data.IP.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error generating MAC for IP", err.Error())
		return
	}

	var apiResp ipMACAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing MAC response", err.Error())
		return
	}

	data.MAC = types.StringValue(apiResp.MAC.MAC)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ipMACResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ipMACResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get(fmt.Sprintf("/ip/%s/mac", data.IP.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MAC for IP", err.Error())
		return
	}

	var apiResp ipMACAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing MAC response", err.Error())
		return
	}

	data.MAC = types.StringValue(apiResp.MAC.MAC)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ipMACResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// IP is ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("Unexpected Update", "IP requires replacement; Update should not be called.")
}

func (r *ipMACResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ipMACResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/ip/%s/mac", data.IP.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error deleting MAC for IP", err.Error())
	}
}

func (r *ipMACResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}
