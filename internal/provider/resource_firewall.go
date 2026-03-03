// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var (
	_ resource.Resource                = &firewallResource{}
	_ resource.ResourceWithImportState = &firewallResource{}
	_ resource.ResourceWithConfigure   = &firewallResource{}
)

type firewallResource struct {
	client *client.Client
}

type firewallRuleModel struct {
	IPVersion types.String `tfsdk:"ip_version"`
	Name      types.String `tfsdk:"name"`
	DstIP     types.String `tfsdk:"dst_ip"`
	SrcIP     types.String `tfsdk:"src_ip"`
	DstPort   types.String `tfsdk:"dst_port"`
	SrcPort   types.String `tfsdk:"src_port"`
	Protocol  types.String `tfsdk:"protocol"`
	TCPFlags  types.String `tfsdk:"tcp_flags"`
	Action    types.String `tfsdk:"action"`
}

type firewallResourceModel struct {
	ServerNumber types.String        `tfsdk:"server_number"`
	ServerIP     types.String        `tfsdk:"server_ip"`
	Status       types.String        `tfsdk:"status"`
	AllowlistHOS types.Bool          `tfsdk:"allowlist_hos"`
	FilterIPv6   types.Bool          `tfsdk:"filter_ipv6"`
	Port         types.String        `tfsdk:"port"`
	InputRules   []firewallRuleModel `tfsdk:"input"`
	OutputRules  []firewallRuleModel `tfsdk:"output"`
}

type firewallAPIResponse struct {
	Firewall firewallAPIModel `json:"firewall"`
}

type firewallAPIModel struct {
	ServerIP     string           `json:"server_ip"`
	ServerNumber int64            `json:"server_number"`
	Status       string           `json:"status"`
	AllowlistHOS bool             `json:"whitelist_hos"`
	FilterIPv6   bool             `json:"filter_ipv6"`
	Port         string           `json:"port"`
	Rules        firewallAPIRules `json:"rules"`
}

type firewallAPIRules struct {
	Input  []firewallAPIRule `json:"input"`
	Output []firewallAPIRule `json:"output"`
}

type firewallAPIRule struct {
	IPVersion string `json:"ip_version"`
	Name      string `json:"name"`
	DstIP     string `json:"dst_ip"`
	SrcIP     string `json:"src_ip"`
	DstPort   string `json:"dst_port"`
	SrcPort   string `json:"src_port"`
	Protocol  string `json:"protocol"`
	TCPFlags  string `json:"tcp_flags"`
	Action    string `json:"action"`
}

func NewFirewallResource() resource.Resource {
	return &firewallResource{}
}

func (r *firewallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

var firewallRuleSchemaAttrs = map[string]schema.Attribute{
	"ip_version": schema.StringAttribute{
		MarkdownDescription: "IP version: `ipv4`, `ipv6`, or empty for both.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"name": schema.StringAttribute{
		MarkdownDescription: "Rule name.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"dst_ip": schema.StringAttribute{
		MarkdownDescription: "Destination IP/subnet in CIDR notation.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"src_ip": schema.StringAttribute{
		MarkdownDescription: "Source IP/subnet in CIDR notation.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"dst_port": schema.StringAttribute{
		MarkdownDescription: "Destination port or range.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"src_port": schema.StringAttribute{
		MarkdownDescription: "Source port or range.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"protocol": schema.StringAttribute{
		MarkdownDescription: "Protocol: `tcp`, `udp`, `icmp`, etc.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"tcp_flags": schema.StringAttribute{
		MarkdownDescription: "TCP flags.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	},
	"action": schema.StringAttribute{
		MarkdownDescription: "Action: `accept` or `discard`.",
		Required:            true,
	},
}

func (r *firewallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages firewall rules for a Hetzner dedicated server. The entire firewall configuration is replaced on each update.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.StringAttribute{
				MarkdownDescription: "Server number. Used as the unique identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: "Server main IP address.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Firewall status: `active` or `disabled`.",
				Required:            true,
			},
			"allowlist_hos": schema.BoolAttribute{
				MarkdownDescription: "Allow Hetzner services (rescue, DHCP, DNS, monitoring).",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"filter_ipv6": schema.BoolAttribute{
				MarkdownDescription: "Whether firewall also filters IPv6 traffic.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"port": schema.StringAttribute{
				MarkdownDescription: "Switch port: `main` or `kvm`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"input": schema.ListNestedAttribute{
				MarkdownDescription: "Incoming traffic rules. Maximum 10 rules. Applied in order.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: firewallRuleSchemaAttrs,
				},
			},
			"output": schema.ListNestedAttribute{
				MarkdownDescription: "Outgoing traffic rules. Maximum 10 rules. Applied in order.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: firewallRuleSchemaAttrs,
				},
			},
		},
	}
}

func (r *firewallResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func firewallRulesToForm(prefix string, rules []firewallRuleModel, form url.Values) {
	for i, rule := range rules {
		idx := strconv.Itoa(i)
		if !rule.IPVersion.IsNull() && !rule.IPVersion.IsUnknown() && rule.IPVersion.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][ip_version]", prefix, idx), rule.IPVersion.ValueString())
		}
		if !rule.Name.IsNull() && !rule.Name.IsUnknown() && rule.Name.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][name]", prefix, idx), rule.Name.ValueString())
		}
		if !rule.DstIP.IsNull() && !rule.DstIP.IsUnknown() && rule.DstIP.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][dst_ip]", prefix, idx), rule.DstIP.ValueString())
		}
		if !rule.SrcIP.IsNull() && !rule.SrcIP.IsUnknown() && rule.SrcIP.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][src_ip]", prefix, idx), rule.SrcIP.ValueString())
		}
		if !rule.DstPort.IsNull() && !rule.DstPort.IsUnknown() && rule.DstPort.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][dst_port]", prefix, idx), rule.DstPort.ValueString())
		}
		if !rule.SrcPort.IsNull() && !rule.SrcPort.IsUnknown() && rule.SrcPort.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][src_port]", prefix, idx), rule.SrcPort.ValueString())
		}
		if !rule.Protocol.IsNull() && !rule.Protocol.IsUnknown() && rule.Protocol.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][protocol]", prefix, idx), rule.Protocol.ValueString())
		}
		if !rule.TCPFlags.IsNull() && !rule.TCPFlags.IsUnknown() && rule.TCPFlags.ValueString() != "" {
			form.Set(fmt.Sprintf("rules[%s][%s][tcp_flags]", prefix, idx), rule.TCPFlags.ValueString())
		}
		form.Set(fmt.Sprintf("rules[%s][%s][action]", prefix, idx), rule.Action.ValueString())
	}
}

// stringOrNull returns a types.StringValue for non-empty strings, types.StringNull for empty.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func apiRulesToModel(apiRules []firewallAPIRule) []firewallRuleModel {
	if len(apiRules) == 0 {
		return nil
	}
	rules := make([]firewallRuleModel, len(apiRules))
	for i, r := range apiRules {
		rules[i] = firewallRuleModel{
			IPVersion: stringOrNull(r.IPVersion),
			Name:      stringOrNull(r.Name),
			DstIP:     stringOrNull(r.DstIP),
			SrcIP:     stringOrNull(r.SrcIP),
			DstPort:   stringOrNull(r.DstPort),
			SrcPort:   stringOrNull(r.SrcPort),
			Protocol:  stringOrNull(r.Protocol),
			TCPFlags:  stringOrNull(r.TCPFlags),
			Action:    types.StringValue(r.Action),
		}
	}
	return rules
}

func (r *firewallResource) setStateFromAPI(data *firewallResourceModel, fw firewallAPIModel) {
	data.ServerNumber = types.StringValue(strconv.FormatInt(fw.ServerNumber, 10))
	data.ServerIP = stringOrNull(fw.ServerIP)
	data.Status = types.StringValue(fw.Status)
	data.AllowlistHOS = types.BoolValue(fw.AllowlistHOS)
	data.FilterIPv6 = types.BoolValue(fw.FilterIPv6)
	data.Port = types.StringValue(fw.Port)
	data.InputRules = apiRulesToModel(fw.Rules.Input)
	data.OutputRules = apiRulesToModel(fw.Rules.Output)
}

func (r *firewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data firewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedInput := data.InputRules
	plannedOutput := data.OutputRules
	plannedStatus := data.Status

	form := url.Values{}
	form.Set("status", data.Status.ValueString())
	if !data.AllowlistHOS.IsNull() && !data.AllowlistHOS.IsUnknown() {
		form.Set("whitelist_hos", strconv.FormatBool(data.AllowlistHOS.ValueBool()))
	}
	if !data.FilterIPv6.IsNull() && !data.FilterIPv6.IsUnknown() {
		form.Set("filter_ipv6", strconv.FormatBool(data.FilterIPv6.ValueBool()))
	}
	if data.InputRules != nil {
		firewallRulesToForm("input", data.InputRules, form)
	}
	if data.OutputRules != nil {
		firewallRulesToForm("output", data.OutputRules, form)
	}

	body, err := r.postWithRetry(ctx, "/firewall/"+data.ServerNumber.ValueString(), form)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall", err.Error())
		return
	}

	var apiResp firewallAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall response", err.Error())
		return
	}

	r.setStateFromAPI(&data, apiResp.Firewall)

	// The API may return "in process" temporarily; use the planned status.
	data.Status = plannedStatus

	if plannedInput == nil {
		data.InputRules = nil
	}
	if plannedOutput == nil {
		data.OutputRules = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data firewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	previousStatus := data.Status
	previousInput := data.InputRules
	previousOutput := data.OutputRules

	body, err := r.client.Get("/firewall/" + data.ServerNumber.ValueString())
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall", err.Error())
		return
	}

	var apiResp firewallAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall response", err.Error())
		return
	}

	r.setStateFromAPI(&data, apiResp.Firewall)

	// The API may return "in process" temporarily during transitions;
	// preserve the previous state's status to avoid drift.
	if apiResp.Firewall.Status == "in process" && !previousStatus.IsNull() {
		data.Status = previousStatus
	}

	// Hetzner auto-manages output rules (Block mail ports, Allow all).
	// If the prior state had nil rules (user didn't configure them),
	// preserve nil to avoid drift. During import, previousStatus is null
	// (only server_number is set), so we populate rules from the API.
	if !previousStatus.IsNull() {
		if previousInput == nil {
			data.InputRules = nil
		}
		if previousOutput == nil {
			data.OutputRules = nil
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data firewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedInput := data.InputRules
	plannedOutput := data.OutputRules
	plannedStatus := data.Status

	form := url.Values{}
	form.Set("status", data.Status.ValueString())
	if !data.AllowlistHOS.IsNull() && !data.AllowlistHOS.IsUnknown() {
		form.Set("whitelist_hos", strconv.FormatBool(data.AllowlistHOS.ValueBool()))
	}
	if !data.FilterIPv6.IsNull() && !data.FilterIPv6.IsUnknown() {
		form.Set("filter_ipv6", strconv.FormatBool(data.FilterIPv6.ValueBool()))
	}
	if data.InputRules != nil {
		firewallRulesToForm("input", data.InputRules, form)
	}
	if data.OutputRules != nil {
		firewallRulesToForm("output", data.OutputRules, form)
	}

	body, err := r.postWithRetry(ctx, "/firewall/"+data.ServerNumber.ValueString(), form)
	if err != nil {
		resp.Diagnostics.AddError("Error updating firewall", err.Error())
		return
	}

	var apiResp firewallAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing firewall response", err.Error())
		return
	}

	r.setStateFromAPI(&data, apiResp.Firewall)

	// The API may return "in process" temporarily; use the planned status.
	data.Status = plannedStatus

	if plannedInput == nil {
		data.InputRules = nil
	}
	if plannedOutput == nil {
		data.OutputRules = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *firewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data firewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverNum := data.ServerNumber.ValueString()

	// Hetzner firewalls are a property of the server and cannot be truly deleted.
	// Best effort: disable the firewall and clear rules.
	form := url.Values{}
	form.Set("status", "disabled")
	form.Set("whitelist_hos", "true")
	// Use retry to handle FIREWALL_IN_PROCESS errors.
	_, _ = r.postWithRetry(ctx, "/firewall/"+serverNum, form)
}

// postWithRetry retries POST requests on FIREWALL_IN_PROCESS (409) errors.
func (r *firewallResource) postWithRetry(ctx context.Context, path string, form url.Values) ([]byte, error) {
	deadline := time.Now().Add(5 * time.Minute)
	retries500 := 0
	for time.Now().Before(deadline) {
		body, err := r.client.Post(path, form)
		if err == nil {
			return body, nil
		}
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 409 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Second):
				continue
			}
		}
		// Retry transient 500 errors a few times.
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 500 && retries500 < 3 {
			retries500++
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Second):
				continue
			}
		}
		return nil, err
	}
	return nil, fmt.Errorf("timed out waiting for firewall to be ready")
}

func (r *firewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath("server_number"), req, resp)
}
