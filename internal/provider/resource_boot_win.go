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
	_ resource.Resource                = &bootWindowsResource{}
	_ resource.ResourceWithConfigure   = &bootWindowsResource{}
	_ resource.ResourceWithImportState = &bootWindowsResource{}
)

type bootWindowsResource struct {
	client *client.Client
}

type bootWindowsResourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	Dist          types.String `tfsdk:"dist"`
	Lang          types.String `tfsdk:"lang"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
	Active        types.Bool   `tfsdk:"active"`
	Password      types.String `tfsdk:"password"`
}

type windowsAPIResponse struct {
	Windows windowsAPIData `json:"windows"`
}

type windowsAPIData struct {
	ServerIP      string      `json:"server_ip"`
	ServerIPv6Net string      `json:"server_ipv6_net"`
	ServerNumber  int         `json:"server_number"`
	Dist          interface{} `json:"dist"`
	Lang          interface{} `json:"lang"`
	Active        bool        `json:"active"`
	Password      *string     `json:"password"`
}

func NewBootWindowsResource() resource.Resource {
	return &bootWindowsResource{}
}

func (r *bootWindowsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boot_windows"
}

func (r *bootWindowsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Windows installation boot configuration for a Hetzner dedicated server.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"dist": schema.StringAttribute{
				MarkdownDescription: "Windows version/distribution.",
				Required:            true,
			},
			"lang": schema.StringAttribute{
				MarkdownDescription: "Language.",
				Required:            true,
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
				MarkdownDescription: "Whether Windows install is currently active.",
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

func (r *bootWindowsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bootWindowsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bootWindowsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := plan.ServerNumber.ValueInt64()
	data := url.Values{}
	data.Set("dist", plan.Dist.ValueString())
	data.Set("lang", plan.Lang.ValueString())

	body, err := r.client.Post(fmt.Sprintf("/boot/%d/windows", serverNum), data)
	if err != nil {
		resp.Diagnostics.AddError("Error activating Windows install", err.Error())
		return
	}

	var apiResp windowsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing Windows response", err.Error())
		return
	}

	plan.ServerIP = types.StringValue(apiResp.Windows.ServerIP)
	plan.ServerIPv6Net = types.StringValue(apiResp.Windows.ServerIPv6Net)
	plan.Active = types.BoolValue(apiResp.Windows.Active)

	if apiResp.Windows.Password != nil {
		plan.Password = types.StringValue(*apiResp.Windows.Password)
	} else {
		plan.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bootWindowsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bootWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := state.ServerNumber.ValueInt64()
	body, err := r.client.Get(fmt.Sprintf("/boot/%d/windows", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error reading Windows boot config", err.Error())
		return
	}

	var apiResp windowsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing Windows response", err.Error())
		return
	}

	state.ServerIP = types.StringValue(apiResp.Windows.ServerIP)
	state.ServerIPv6Net = types.StringValue(apiResp.Windows.ServerIPv6Net)
	state.Active = types.BoolValue(apiResp.Windows.Active)

	if apiResp.Windows.Active {
		if distStr, ok := apiResp.Windows.Dist.(string); ok {
			state.Dist = types.StringValue(distStr)
		}
		if langStr, ok := apiResp.Windows.Lang.(string); ok {
			state.Lang = types.StringValue(langStr)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bootWindowsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bootWindowsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := plan.ServerNumber.ValueInt64()

	_, err := r.client.Delete(fmt.Sprintf("/boot/%d/windows", serverNum))
	if err != nil {
		resp.Diagnostics.AddError("Error deactivating Windows install", err.Error())
		return
	}

	data := url.Values{}
	data.Set("dist", plan.Dist.ValueString())
	data.Set("lang", plan.Lang.ValueString())

	body, err := r.client.Post(fmt.Sprintf("/boot/%d/windows", serverNum), data)
	if err != nil {
		resp.Diagnostics.AddError("Error reactivating Windows install", err.Error())
		return
	}

	var apiResp windowsAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing Windows response", err.Error())
		return
	}

	plan.ServerIP = types.StringValue(apiResp.Windows.ServerIP)
	plan.ServerIPv6Net = types.StringValue(apiResp.Windows.ServerIPv6Net)
	plan.Active = types.BoolValue(apiResp.Windows.Active)

	if apiResp.Windows.Password != nil {
		plan.Password = types.StringValue(*apiResp.Windows.Password)
	} else {
		plan.Password = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bootWindowsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bootWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/boot/%d/windows", state.ServerNumber.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error deactivating Windows install", err.Error())
		return
	}
}

func (r *bootWindowsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serverNum, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected numeric server_number")
		return
	}

	var state bootWindowsResourceModel
	state.ServerNumber = types.Int64Value(serverNum)
	state.Dist = types.StringNull()
	state.Lang = types.StringNull()
	state.ServerIP = types.StringNull()
	state.ServerIPv6Net = types.StringNull()
	state.Active = types.BoolNull()
	state.Password = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
