// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &serverOrderResource{}
	_ resource.ResourceWithConfigure   = &serverOrderResource{}
	_ resource.ResourceWithImportState = &serverOrderResource{}
)

func NewServerOrderResource() resource.Resource {
	return &serverOrderResource{}
}

type serverOrderResource struct {
	client *client.Client
}

type serverOrderResourceModel struct {
	ProductID      types.String `tfsdk:"product_id"`
	Source         types.String `tfsdk:"source"`
	AuthorizedKeys types.List   `tfsdk:"authorized_keys"`
	Addons         types.List   `tfsdk:"addons"`
	Location       types.String `tfsdk:"location"`
	Dist           types.String `tfsdk:"dist"`
	Lang           types.String `tfsdk:"lang"`
	Test           types.Bool   `tfsdk:"test"`
	TransactionID  types.String `tfsdk:"transaction_id"`
	ServerNumber   types.Int64  `tfsdk:"server_number"`
	ServerIP       types.String `tfsdk:"server_ip"`
	ServerIPv6     types.String `tfsdk:"server_ipv6_net"`
	ServerName     types.String `tfsdk:"server_name"`
	Product        types.String `tfsdk:"product"`
	DC             types.String `tfsdk:"dc"`
	Traffic        types.String `tfsdk:"traffic"`
	Status         types.String `tfsdk:"status"`
	Cancelled      types.Bool   `tfsdk:"cancelled"`
	PaidUntil      types.String `tfsdk:"paid_until"`
}

// orderTransactionAPIResponse represents the API response for server order transactions.
type orderTransactionAPIResponse struct {
	Transaction orderTransactionAPI `json:"transaction"`
}

type orderTransactionAPI struct {
	ID           string `json:"id"`
	ServerNumber *int   `json:"server_number"`
	Status       string `json:"status"`
}

func (r *serverOrderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_order"
}

func (r *serverOrderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Orders a Hetzner dedicated server from the standard catalog or server market (auction). The server is cancelled on destroy.",
		Attributes: map[string]schema.Attribute{
			"product_id": schema.StringAttribute{
				MarkdownDescription: "Product ID to order.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "Order source: \"market\" (auction) or \"standard\" (catalog). Defaults to \"market\".",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("market"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authorized_keys": schema.ListAttribute{
				MarkdownDescription: "List of SSH key fingerprints to authorize.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"addons": schema.ListAttribute{
				MarkdownDescription: "List of addon product IDs (e.g. \"primary_ipv4\").",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Preferred data center location (standard orders only).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dist": schema.StringAttribute{
				MarkdownDescription: "Distribution to install.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lang": schema.StringAttribute{
				MarkdownDescription: "Language for the installation.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"test": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a test order. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"transaction_id": schema.StringAttribute{
				MarkdownDescription: "The order transaction ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The assigned server number.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "The main IPv4 address of the server.",
				Computed:            true,
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "The main IPv6 network of the server.",
				Computed:            true,
			},
			"server_name": schema.StringAttribute{
				MarkdownDescription: "The server name.",
				Computed:            true,
			},
			"product": schema.StringAttribute{
				MarkdownDescription: "The product name.",
				Computed:            true,
			},
			"dc": schema.StringAttribute{
				MarkdownDescription: "The data center.",
				Computed:            true,
			},
			"traffic": schema.StringAttribute{
				MarkdownDescription: "Free traffic quota.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Server status.",
				Computed:            true,
			},
			"cancelled": schema.BoolAttribute{
				MarkdownDescription: "Whether the server has been cancelled.",
				Computed:            true,
			},
			"paid_until": schema.StringAttribute{
				MarkdownDescription: "Date the server is paid until.",
				Computed:            true,
			},
		},
	}
}

func (r *serverOrderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverOrderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build form params.
	form := url.Values{}
	form.Set("product_id", data.ProductID.ValueString())

	if !data.AuthorizedKeys.IsNull() {
		var keys []string
		resp.Diagnostics.Append(data.AuthorizedKeys.ElementsAs(ctx, &keys, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, k := range keys {
			form.Add("authorized_key[]", k)
		}
	}

	if !data.Addons.IsNull() {
		var addons []string
		resp.Diagnostics.Append(data.Addons.ElementsAs(ctx, &addons, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, a := range addons {
			form.Add("addon[]", a)
		}
	}

	if !data.Location.IsNull() {
		form.Set("location", data.Location.ValueString())
	}
	if !data.Dist.IsNull() {
		form.Set("dist", data.Dist.ValueString())
	}
	if !data.Lang.IsNull() {
		form.Set("lang", data.Lang.ValueString())
	}
	if data.Test.ValueBool() {
		form.Set("test", "true")
	}

	// Determine endpoint based on source.
	endpoint := "/order/server_market/transaction"
	txnPollBase := "/order/server_market/transaction"
	if data.Source.ValueString() == "standard" {
		endpoint = "/order/server/transaction"
		txnPollBase = "/order/server/transaction"
	}

	// Place the order.
	body, err := r.client.PostWithContext(ctx, endpoint, form)
	if err != nil {
		resp.Diagnostics.AddError("Error ordering server", err.Error())
		return
	}

	var orderResp orderTransactionAPIResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		resp.Diagnostics.AddError("Error parsing order response", err.Error())
		return
	}

	data.TransactionID = types.StringValue(orderResp.Transaction.ID)

	// Wait for server_number assignment.
	serverNumber := 0
	if orderResp.Transaction.ServerNumber != nil && *orderResp.Transaction.ServerNumber != 0 {
		serverNumber = *orderResp.Transaction.ServerNumber
	} else {
		sn, diags := r.pollTransactionForServerNumber(ctx, txnPollBase, orderResp.Transaction.ID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		serverNumber = sn
	}

	data.ServerNumber = types.Int64Value(int64(serverNumber))

	// Wait for server to be ready.
	r.pollServerReady(ctx, serverNumber, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read full server details.
	r.readServerOrder(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data serverOrderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readServerOrder(&data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *serverOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("Unexpected Update", "All attributes require replacement; Update should not be called.")
}

func (r *serverOrderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data serverOrderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("cancellation_date", "now")

	_, err := r.client.PostWithContext(ctx, fmt.Sprintf("/server/%d/cancellation", data.ServerNumber.ValueInt64()), form)
	if err != nil {
		// 409 means already cancelled - that's fine.
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 409 {
			return
		}
		resp.Diagnostics.AddError("Error cancelling server", err.Error())
	}
}

func (r *serverOrderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "server_number must be a numeric value")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_number"), id)...)
}

// readServerOrder fetches server details from GET /server/{number} and populates computed fields.
func (r *serverOrderResource) readServerOrder(data *serverOrderResourceModel, diags *diag.Diagnostics) {
	body, err := r.client.Get(fmt.Sprintf("/server/%d", data.ServerNumber.ValueInt64()))
	if err != nil {
		diags.AddError("Error reading server", err.Error())
		return
	}

	var apiResp serverDetailAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		diags.AddError("Error parsing server response", err.Error())
		return
	}

	s := apiResp.Server
	data.ServerNumber = types.Int64Value(int64(s.ServerNumber))
	data.ServerIP = stringOrNull(s.ServerIP)
	data.ServerIPv6 = stringOrNull(s.ServerIPv6)
	data.ServerName = types.StringValue(s.ServerName)
	data.Product = types.StringValue(s.Product)
	data.DC = types.StringValue(s.DC)
	data.Traffic = types.StringValue(s.Traffic)
	data.Status = types.StringValue(s.Status)
	data.Cancelled = types.BoolValue(s.Cancelled)
	data.PaidUntil = types.StringValue(s.PaidUntil)
}

// pollTransactionForServerNumber polls the transaction endpoint until server_number is assigned.
func (r *serverOrderResource) pollTransactionForServerNumber(ctx context.Context, basePath, txnID string) (int, diag.Diagnostics) {
	var diags diag.Diagnostics
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	deadline := time.After(20 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			diags.AddError("Context cancelled", "Timed out waiting for server_number assignment")
			return 0, diags
		case <-deadline:
			diags.AddError("Timeout", fmt.Sprintf("Timed out after 20 minutes waiting for server_number from transaction %s", txnID))
			return 0, diags
		case <-ticker.C:
			body, err := r.client.GetWithContext(ctx, fmt.Sprintf("%s/%s", basePath, txnID))
			if err != nil {
				continue
			}
			var txnResp orderTransactionAPIResponse
			if err := json.Unmarshal(body, &txnResp); err != nil {
				continue
			}
			if txnResp.Transaction.ServerNumber != nil && *txnResp.Transaction.ServerNumber != 0 {
				return *txnResp.Transaction.ServerNumber, diags
			}
		}
	}
}

// pollServerReady polls GET /server/{number} until the server status is "ready".
func (r *serverOrderResource) pollServerReady(ctx context.Context, serverNumber int, diags *diag.Diagnostics) {
	checkReady := func() bool {
		body, err := r.client.GetWithContext(ctx, fmt.Sprintf("/server/%d", serverNumber))
		if err != nil {
			return false
		}
		var resp serverDetailAPIResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return false
		}
		return resp.Server.Status == "ready"
	}

	// Check immediately before starting the ticker.
	if checkReady() {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	deadline := time.After(30 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			diags.AddError("Context cancelled", "Timed out waiting for server to be ready")
			return
		case <-deadline:
			diags.AddError("Timeout", fmt.Sprintf("Timed out after 30 minutes waiting for server %d to be ready", serverNumber))
			return
		case <-ticker.C:
			if checkReady() {
				return
			}
		}
	}
}
