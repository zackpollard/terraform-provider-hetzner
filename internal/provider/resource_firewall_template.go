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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &firewallTemplateResource{}
	_ resource.ResourceWithImportState = &firewallTemplateResource{}
	_ resource.ResourceWithConfigure   = &firewallTemplateResource{}
)

type firewallTemplateResource struct {
	client *client.Client
}

type firewallTemplateResourceModel struct {
	ID           types.String        `tfsdk:"id"`
	Name         types.String        `tfsdk:"name"`
	FilterIPv6   types.Bool          `tfsdk:"filter_ipv6"`
	AllowlistHOS types.Bool          `tfsdk:"allowlist_hos"`
	IsDefault    types.Bool          `tfsdk:"is_default"`
	InputRules   []firewallRuleModel `tfsdk:"input"`
	OutputRules  []firewallRuleModel `tfsdk:"output"`
}

type firewallTemplateAPIResponse struct {
	FirewallTemplate firewallTemplateAPIModel `json:"firewall_template"`
}

type firewallTemplateAPIModel struct {
	ID           int64            `json:"id"`
	Name         string           `json:"name"`
	FilterIPv6   bool             `json:"filter_ipv6"`
	AllowlistHOS bool             `json:"allowlist_hos"`
	IsDefault    bool             `json:"is_default"`
	Rules        firewallAPIRules `json:"rules"`
}

func NewFirewallTemplateResource() resource.Resource {
	return &firewallTemplateResource{}
}

func (r *firewallTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_template"
}

func (r *firewallTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a reusable firewall template in the Hetzner Robot API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Template ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name.",
				Required:            true,
			},
			"filter_ipv6": schema.BoolAttribute{
				MarkdownDescription: "Whether template filters IPv6 traffic.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"allowlist_hos": schema.BoolAttribute{
				MarkdownDescription: "Allow Hetzner services.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the default template.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"input": schema.ListNestedAttribute{
				MarkdownDescription: "Incoming traffic rules.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: firewallRuleSchemaAttrs,
				},
			},
			"output": schema.ListNestedAttribute{
				MarkdownDescription: "Outgoing traffic rules.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: firewallRuleSchemaAttrs,
				},
			},
		},
	}
}

func (r *firewallTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *firewallTemplateResource) templateToForm(data *firewallTemplateResourceModel) url.Values {
	form := url.Values{}
	form.Set("name", data.Name.ValueString())
	if !data.FilterIPv6.IsNull() && !data.FilterIPv6.IsUnknown() {
		form.Set("filter_ipv6", strconv.FormatBool(data.FilterIPv6.ValueBool()))
	}
	if !data.AllowlistHOS.IsNull() && !data.AllowlistHOS.IsUnknown() {
		form.Set("allowlist_hos", strconv.FormatBool(data.AllowlistHOS.ValueBool()))
	}
	if !data.IsDefault.IsNull() && !data.IsDefault.IsUnknown() {
		form.Set("is_default", strconv.FormatBool(data.IsDefault.ValueBool()))
	}
	if data.InputRules != nil {
		firewallRulesToForm("input", data.InputRules, form)
	}
	if data.OutputRules != nil {
		firewallRulesToForm("output", data.OutputRules, form)
	}
	return form
}

func (r *firewallTemplateResource) setStateFromAPI(data *firewallTemplateResourceModel, tmpl firewallTemplateAPIModel) {
	data.ID = types.StringValue(strconv.FormatInt(tmpl.ID, 10))
	data.Name = types.StringValue(tmpl.Name)
	data.FilterIPv6 = types.BoolValue(tmpl.FilterIPv6)
	data.AllowlistHOS = types.BoolValue(tmpl.AllowlistHOS)
	data.IsDefault = types.BoolValue(tmpl.IsDefault)
	data.InputRules = apiRulesToModel(tmpl.Rules.Input)
	data.OutputRules = apiRulesToModel(tmpl.Rules.Output)
}

func (r *firewallTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data firewallTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := r.templateToForm(&data)

	body, err := r.client.Post("/firewall/template", form)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall template", err.Error())
		return
	}

	var apiResp firewallTemplateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall template response", err.Error())
		return
	}

	r.setStateFromAPI(&data, apiResp.FirewallTemplate)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data firewallTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.Get("/firewall/template/" + data.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall template", err.Error())
		return
	}

	var apiResp firewallTemplateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall template response", err.Error())
		return
	}

	r.setStateFromAPI(&data, apiResp.FirewallTemplate)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data firewallTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state firewallTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := r.templateToForm(&data)

	body, err := r.client.Post("/firewall/template/"+state.ID.ValueString(), form)
	if err != nil {
		resp.Diagnostics.AddError("Error updating firewall template", err.Error())
		return
	}

	var apiResp firewallTemplateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall template response", err.Error())
		return
	}

	r.setStateFromAPI(&data, apiResp.FirewallTemplate)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data firewallTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete("/firewall/template/" + data.ID.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall template", err.Error())
	}
}

func (r *firewallTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath("id"), req, resp)
}
