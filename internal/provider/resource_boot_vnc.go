// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &bootVNCResource{}
	_ resource.ResourceWithConfigure   = &bootVNCResource{}
	_ resource.ResourceWithImportState = &bootVNCResource{}
)

type bootVNCResource struct {
	client *client.Client
}

type bootVNCResourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	Dist          types.String `tfsdk:"dist"`
	Lang          types.String `tfsdk:"lang"`
	Arch          types.Int64  `tfsdk:"arch"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
	Active        types.Bool   `tfsdk:"active"`
	Password      types.String `tfsdk:"password"`
}

type vncAPIResponse struct {
	VNC vncAPIData `json:"vnc"`
}

type vncAPIData struct {
	ServerIP      string      `json:"server_ip"`
	ServerIPv6Net string      `json:"server_ipv6_net"`
	ServerNumber  int         `json:"server_number"`
	Dist          interface{} `json:"dist"`
	Lang          interface{} `json:"lang"`
	Active        bool        `json:"active"`
	Password      *string     `json:"password"`
}

func NewBootVNCResource() resource.Resource {
	return &bootVNCResource{}
}

func (r *bootVNCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boot_vnc"
}

func (r *bootVNCResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the VNC installation boot configuration for a Hetzner dedicated server.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"dist": schema.StringAttribute{
				MarkdownDescription: "Distribution for VNC install.",
				Required:            true,
			},
			"lang": schema.StringAttribute{
				MarkdownDescription: "Language.",
				Required:            true,
			},
			"arch": schema.Int64Attribute{
				MarkdownDescription: "Architecture (64 or 32). Defaults to 64.",
				Optional:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Server main IPv4 address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_ipv6_net": schema.StringAttribute{
				MarkdownDescription: "Server IPv6 network.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether VNC install is currently active.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Generated password. Only available on activation.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *bootVNCResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bootVNCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bootVNCResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := plan.ServerNumber.ValueInt64()
	data := url.Values{}
	data.Set("dist", plan.Dist.ValueString())
	data.Set("lang", plan.Lang.ValueString())

	if !plan.Arch.IsNull() && !plan.Arch.IsUnknown() {
		data.Set("arch", strconv.FormatInt(plan.Arch.ValueInt64(), 10))
	}

	body, err := r.client.Post(fmt.Sprintf("/boot/%d/vnc", serverNum), data)
	if err != nil {
		resp.Diagnostics.AddError("Error activating VNC install", err.Error())
		return
	}

	var apiResp vncAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing VNC response", err.Error())
		return
	}

	plan.ServerIP = types.StringValue(apiResp.VNC.ServerIP)
	plan.ServerIPv6Net = types.StringValue(apiResp.VNC.ServerIPv6Net)
	plan.Active = types.BoolValue(apiResp.VNC.Active)

	if apiResp.VNC.Password != nil {
		plan.Password = types.StringValue(*apiResp.VNC.Password)
	} else {
		plan.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bootVNCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bootVNCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := state.ServerNumber.ValueInt64()
	body, err := r.client.Get(fmt.Sprintf("/boot/%d/vnc", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading VNC boot config", err.Error())
		return
	}

	var apiResp vncAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing VNC response", err.Error())
		return
	}

	state.ServerIP = types.StringValue(apiResp.VNC.ServerIP)
	state.ServerIPv6Net = types.StringValue(apiResp.VNC.ServerIPv6Net)
	state.Active = types.BoolValue(apiResp.VNC.Active)

	if apiResp.VNC.Active {
		if distStr, ok := apiResp.VNC.Dist.(string); ok {
			state.Dist = types.StringValue(distStr)
		}
		if langStr, ok := apiResp.VNC.Lang.(string); ok {
			state.Lang = types.StringValue(langStr)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bootVNCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bootVNCResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := plan.ServerNumber.ValueInt64()

	_, err := r.client.Delete(fmt.Sprintf("/boot/%d/vnc", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error deactivating VNC install", err.Error())
		return
	}

	data := url.Values{}
	data.Set("dist", plan.Dist.ValueString())
	data.Set("lang", plan.Lang.ValueString())

	if !plan.Arch.IsNull() && !plan.Arch.IsUnknown() {
		data.Set("arch", strconv.FormatInt(plan.Arch.ValueInt64(), 10))
	}

	body, err := r.client.Post(fmt.Sprintf("/boot/%d/vnc", serverNum), data)
	if err != nil {
		resp.Diagnostics.AddError("Error reactivating VNC install", err.Error())
		return
	}

	var apiResp vncAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing VNC response", err.Error())
		return
	}

	plan.ServerIP = types.StringValue(apiResp.VNC.ServerIP)
	plan.ServerIPv6Net = types.StringValue(apiResp.VNC.ServerIPv6Net)
	plan.Active = types.BoolValue(apiResp.VNC.Active)

	if apiResp.VNC.Password != nil {
		plan.Password = types.StringValue(*apiResp.VNC.Password)
	} else {
		plan.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bootVNCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bootVNCResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/boot/%d/vnc", state.ServerNumber.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error deactivating VNC install", err.Error())
		return
	}
}

func (r *bootVNCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serverNum, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected numeric server_number")
		return
	}

	var state bootVNCResourceModel
	state.ServerNumber = types.Int64Value(serverNum)
	state.Dist = types.StringNull()
	state.Lang = types.StringNull()
	state.Arch = types.Int64Null()
	state.ServerIP = types.StringNull()
	state.ServerIPv6Net = types.StringNull()
	state.Active = types.BoolNull()
	state.Password = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
