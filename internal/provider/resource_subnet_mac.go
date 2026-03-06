// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &subnetMACResource{}
	_ resource.ResourceWithImportState = &subnetMACResource{}
)

func NewSubnetMACResource() resource.Resource {
	return &subnetMACResource{}
}

type subnetMACResource struct {
	client *client.Client
}

type subnetMACResourceModel struct {
	IP  types.String `tfsdk:"ip"`
	MAC types.String `tfsdk:"mac"`
}

type subnetMACAPIResponse struct {
	MAC subnetMACAPIData `json:"mac"`
}

type subnetMACAPIData struct {
	IP   string `json:"ip"`
	Mask string `json:"mask"`
	MAC  string `json:"mac"`
}

func (r *subnetMACResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_mac"
}

func (r *subnetMACResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a separate MAC address for a Hetzner subnet.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The subnet IP address.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac": schema.StringAttribute{
				MarkdownDescription: "The MAC address to assign. Must be one of the available MAC addresses from the subnet.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *subnetMACResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subnetMACResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data subnetMACResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("mac", data.MAC.ValueString())

	body, err := r.client.Put(fmt.Sprintf("/subnet/%s/mac", data.IP.ValueString()), form)
	if err != nil {
		resp.Diagnostics.AddError("Error setting MAC for subnet", err.Error())
		return
	}

	var apiResp subnetMACAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing MAC response", err.Error())
		return
	}

	data.MAC = types.StringValue(apiResp.MAC.MAC)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetMACResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subnetMACResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get(fmt.Sprintf("/subnet/%s/mac", data.IP.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading MAC for subnet", err.Error())
		return
	}

	var apiResp subnetMACAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing MAC response", err.Error())
		return
	}

	if apiResp.MAC.MAC == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	data.MAC = types.StringValue(apiResp.MAC.MAC)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetMACResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("Unexpected Update", "All attributes require replacement; Update should not be called.")
}

func (r *subnetMACResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data subnetMACResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/subnet/%s/mac", data.IP.ValueString()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error deleting MAC for subnet", err.Error())
	}
}

func (r *subnetMACResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}
