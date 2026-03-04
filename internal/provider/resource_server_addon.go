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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &serverAddonResource{}
	_ resource.ResourceWithConfigure   = &serverAddonResource{}
	_ resource.ResourceWithImportState = &serverAddonResource{}
)

func NewServerAddonResource() resource.Resource {
	return &serverAddonResource{}
}

type serverAddonResource struct {
	client *client.Client
}

type serverAddonResourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ProductID     types.String `tfsdk:"product_id"`
	TransactionID types.String `tfsdk:"transaction_id"`
	Status        types.String `tfsdk:"status"`
}

// addonTransactionAPIResponse represents the API response for addon order transactions.
type addonTransactionAPIResponse struct {
	Transaction addonTransactionAPI `json:"transaction"`
}

type addonTransactionAPI struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	ServerNumber *int   `json:"server_number"`
	ProductID    string `json:"product_id,omitempty"`
}

func (r *serverAddonResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_addon"
}

func (r *serverAddonResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Orders an addon for a Hetzner dedicated server (e.g. primary_ipv4).",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number to add the addon to.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"product_id": schema.StringAttribute{
				MarkdownDescription: "The addon product ID (e.g. \"primary_ipv4\").",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"transaction_id": schema.StringAttribute{
				MarkdownDescription: "The order transaction ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The transaction status.",
				Computed:            true,
			},
		},
	}
}

func (r *serverAddonResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverAddonResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverAddonResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("product_id", data.ProductID.ValueString())
	form.Set("server_number", fmt.Sprintf("%d", data.ServerNumber.ValueInt64()))

	body, err := r.client.PostWithContext(ctx, "/order/server_addon/transaction", form)
	if err != nil {
		resp.Diagnostics.AddError("Error ordering server addon", err.Error())
		return
	}

	var apiResp addonTransactionAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing addon order response", err.Error())
		return
	}

	data.TransactionID = types.StringValue(apiResp.Transaction.ID)
	data.Status = types.StringValue(apiResp.Transaction.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverAddonResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverAddonResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.GetWithContext(ctx, fmt.Sprintf("/order/server_addon/transaction/%s", data.TransactionID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading addon transaction", err.Error())
		return
	}

	var apiResp addonTransactionAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing addon transaction response", err.Error())
		return
	}

	data.Status = types.StringValue(apiResp.Transaction.Status)
	if apiResp.Transaction.ServerNumber != nil {
		data.ServerNumber = types.Int64Value(int64(*apiResp.Transaction.ServerNumber))
	}
	if apiResp.Transaction.ProductID != "" {
		data.ProductID = types.StringValue(apiResp.Transaction.ProductID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverAddonResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("Unexpected Update", "All attributes require replacement; Update should not be called.")
}

func (r *serverAddonResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Addons cannot be un-ordered. Just remove from state.
}

func (r *serverAddonResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("transaction_id"), req.ID)...)
}
