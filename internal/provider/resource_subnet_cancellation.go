// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
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
	_ resource.Resource                = &subnetCancellationResource{}
	_ resource.ResourceWithImportState = &subnetCancellationResource{}
)

func NewSubnetCancellationResource() resource.Resource {
	return &subnetCancellationResource{}
}

type subnetCancellationResource struct {
	client *client.Client
}

type subnetCancellationResourceModel struct {
	IP                       types.String `tfsdk:"ip"`
	CancellationDate         types.String `tfsdk:"cancellation_date"`
	EarliestCancellationDate types.String `tfsdk:"earliest_cancellation_date"`
	Cancelled                types.Bool   `tfsdk:"cancelled"`
	ServerNumber             types.String `tfsdk:"server_number"`
}

type subnetCancellationAPIResponse struct {
	Cancellation subnetCancellationAPI `json:"cancellation"`
}

type subnetCancellationAPI struct {
	IP                       string  `json:"ip"`
	ServerNumber             string  `json:"server_number"`
	EarliestCancellationDate string  `json:"earliest_cancellation_date"`
	Cancelled                bool    `json:"cancelled"`
	CancellationDate         *string `json:"cancellation_date"`
}

func (r *subnetCancellationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_cancellation"
}

func (r *subnetCancellationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages cancellation for a Hetzner subnet. Creating this resource cancels the subnet; destroying it withdraws the cancellation.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.StringAttribute{
				MarkdownDescription: "The subnet IP address to cancel.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cancellation_date": schema.StringAttribute{
				MarkdownDescription: "Cancellation date in yyyy-MM-dd format, or \"now\" for immediate cancellation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"earliest_cancellation_date": schema.StringAttribute{
				MarkdownDescription: "Earliest possible cancellation date.",
				Computed:            true,
			},
			"cancelled": schema.BoolAttribute{
				MarkdownDescription: "Whether the subnet is cancelled.",
				Computed:            true,
			},
			"server_number": schema.StringAttribute{
				MarkdownDescription: "The server number the subnet is assigned to.",
				Computed:            true,
			},
		},
	}
}

func (r *subnetCancellationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subnetCancellationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data subnetCancellationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("cancellation_date", data.CancellationDate.ValueString())

	body, err := r.client.Post(fmt.Sprintf("/subnet/%s/cancellation", data.IP.ValueString()), form)
	if err != nil {
		resp.Diagnostics.AddError("Error cancelling subnet", err.Error())
		return
	}

	var apiResp subnetCancellationAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing cancellation response", err.Error())
		return
	}

	r.mapAPIToModel(&data, &apiResp.Cancellation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetCancellationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data subnetCancellationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get(fmt.Sprintf("/subnet/%s/cancellation", data.IP.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading subnet cancellation", err.Error())
		return
	}

	var apiResp subnetCancellationAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing cancellation response", err.Error())
		return
	}

	if !apiResp.Cancellation.Cancelled {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapAPIToModel(&data, &apiResp.Cancellation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *subnetCancellationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unexpected Update", "All attributes require replacement; Update should not be called.")
}

func (r *subnetCancellationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data subnetCancellationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/subnet/%s/cancellation", data.IP.ValueString()))
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error withdrawing subnet cancellation", err.Error())
	}
}

func (r *subnetCancellationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}

func (r *subnetCancellationResource) mapAPIToModel(data *subnetCancellationResourceModel, api *subnetCancellationAPI) {
	data.EarliestCancellationDate = types.StringValue(api.EarliestCancellationDate)
	data.Cancelled = types.BoolValue(api.Cancelled)
	data.ServerNumber = types.StringValue(api.ServerNumber)
	if api.CancellationDate != nil {
		data.CancellationDate = types.StringValue(*api.CancellationDate)
	}
}
