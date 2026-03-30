// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ datasource.DataSource = &trafficDataSource{}

func NewTrafficDataSource() datasource.DataSource {
	return &trafficDataSource{}
}

type trafficDataSource struct {
	client *client.Client
}

type trafficDataSourceModel struct {
	IPs          []types.String      `tfsdk:"ips"`
	Subnets      []types.String      `tfsdk:"subnets"`
	From         types.String        `tfsdk:"from"`
	To           types.String        `tfsdk:"to"`
	Type         types.String        `tfsdk:"type"`
	SingleValues types.Bool          `tfsdk:"single_values"`
	Data         []trafficEntryModel `tfsdk:"data"`
}

type trafficEntryModel struct {
	IP  types.String  `tfsdk:"ip"`
	In  types.Float64 `tfsdk:"in"`
	Out types.Float64 `tfsdk:"out"`
	Sum types.Float64 `tfsdk:"sum"`
}

type trafficAPIResponse struct {
	Traffic trafficAPIData `json:"traffic"`
}

type trafficAPIData struct {
	Type string          `json:"type"`
	From string          `json:"from"`
	To   string          `json:"to"`
	Data json.RawMessage `json:"data"`
}

type trafficAPIEntry struct {
	In  float64 `json:"in"`
	Out float64 `json:"out"`
	Sum float64 `json:"sum"`
}

func (d *trafficDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic"
}

func (d *trafficDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Query traffic statistics for Hetzner IP addresses and subnets.",
		Attributes: map[string]schema.Attribute{
			"ips": schema.ListAttribute{
				MarkdownDescription: "List of IP addresses to query traffic for.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"subnets": schema.ListAttribute{
				MarkdownDescription: "List of subnet IPs to query traffic for.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "Start date/time. Format depends on type: day=YYYY-MM-DDTHH, month=YYYY-MM-DD, year=YYYY-MM.",
				Required:            true,
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "End date/time. Format depends on type: day=YYYY-MM-DDTHH, month=YYYY-MM-DD, year=YYYY-MM.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Query type: \"day\", \"month\", or \"year\".",
				Required:            true,
			},
			"single_values": schema.BoolAttribute{
				MarkdownDescription: "Whether to return grouped data.",
				Optional:            true,
			},
			"data": schema.ListNestedAttribute{
				MarkdownDescription: "Aggregated traffic data per IP/subnet.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							MarkdownDescription: "The IP address or subnet.",
							Computed:            true,
						},
						"in": schema.Float64Attribute{
							MarkdownDescription: "Inbound traffic in GB.",
							Computed:            true,
						},
						"out": schema.Float64Attribute{
							MarkdownDescription: "Outbound traffic in GB.",
							Computed:            true,
						},
						"sum": schema.Float64Attribute{
							MarkdownDescription: "Total traffic in GB.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *trafficDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *trafficDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data trafficDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	form := url.Values{}
	form.Set("from", data.From.ValueString())
	form.Set("to", data.To.ValueString())
	form.Set("type", data.Type.ValueString())

	for _, ip := range data.IPs {
		form.Add("ip[]", ip.ValueString())
	}
	for _, s := range data.Subnets {
		form.Add("subnet[]", s.ValueString())
	}
	if !data.SingleValues.IsNull() && data.SingleValues.ValueBool() {
		form.Set("single_values", "true")
	}

	body, err := d.client.Post("/traffic", form)
	if err != nil {
		resp.Diagnostics.AddError("Error querying traffic", err.Error())
		return
	}

	var apiResp trafficAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing traffic response", err.Error())
		return
	}

	// The data field can be a map[string]array or an empty array when no data.
	dataMap, err := parseTrafficData(apiResp.Traffic.Data)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing traffic data", err.Error())
		return
	}

	data.Data = make([]trafficEntryModel, 0, len(dataMap))
	for ip, entry := range dataMap {
		data.Data = append(data.Data, trafficEntryModel{
			IP:  types.StringValue(ip),
			In:  types.Float64Value(entry.In),
			Out: types.Float64Value(entry.Out),
			Sum: types.Float64Value(entry.Sum),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// parseTrafficData handles the traffic API response data field, which can be:
// - A map of IP -> array of traffic entries (normal case)
// - An empty array [] (no data).
func parseTrafficData(raw json.RawMessage) (map[string]trafficAPIEntry, error) {
	result := make(map[string]trafficAPIEntry)

	// Try map format first: {"ip": [{in, out, sum}, ...]}
	var dataMap map[string]json.RawMessage
	if json.Unmarshal(raw, &dataMap) == nil {
		for ip, ipRaw := range dataMap {
			result[ip] = parseTrafficEntry(ipRaw)
		}
		return result, nil
	}

	// Empty array [] means no data.
	var arr []json.RawMessage
	if json.Unmarshal(raw, &arr) == nil && len(arr) == 0 {
		return result, nil
	}

	return result, nil
}

// parseTrafficEntry handles both single-object and array-of-objects responses.
// For arrays, it sums all entries.
func parseTrafficEntry(raw json.RawMessage) trafficAPIEntry {
	// Try single object first.
	var single trafficAPIEntry
	if json.Unmarshal(raw, &single) == nil && (single.In != 0 || single.Out != 0 || single.Sum != 0) {
		return single
	}

	// Try array of objects.
	var entries []trafficAPIEntry
	if json.Unmarshal(raw, &entries) == nil {
		var total trafficAPIEntry
		for _, e := range entries {
			total.In += e.In
			total.Out += e.Out
			total.Sum += e.Sum
		}
		return total
	}

	return trafficAPIEntry{}
}
