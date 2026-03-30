// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ datasource.DataSource = &serverAddonsDataSource{}

func NewServerAddonsDataSource() datasource.DataSource {
	return &serverAddonsDataSource{}
}

type serverAddonsDataSource struct {
	client *client.Client
}

type serverAddonsModel struct {
	ServerNumber types.Int64        `tfsdk:"server_number"`
	Addons       []serverAddonModel `tfsdk:"addons"`
}

type serverAddonModel struct {
	ID    types.String              `tfsdk:"id"`
	Name  types.String              `tfsdk:"name"`
	Type  types.String              `tfsdk:"type"`
	Price []serverOrderProductPrice `tfsdk:"price"`
}

// API response types

type serverAddonAPIResponse struct {
	Addon serverAddonAPI `json:"addon"`
}

type serverAddonAPI struct {
	ID    string                       `json:"id"`
	Name  string                       `json:"name"`
	Type  string                       `json:"type"`
	Price []serverOrderProductPriceAPI `json:"price"`
}

func (d *serverAddonsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_addons"
}

func (d *serverAddonsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available addons for a Hetzner dedicated server.",
		Attributes: map[string]schema.Attribute{
			"server_number": schema.Int64Attribute{
				MarkdownDescription: "The server number to list addons for.",
				Required:            true,
			},
			"addons": schema.ListNestedAttribute{
				MarkdownDescription: "List of available addons.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Addon ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Addon name.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Addon type.",
							Computed:            true,
						},
						"price": schema.ListNestedAttribute{
							MarkdownDescription: "Pricing information per location.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: priceSchemaAttrs,
							},
						},
					},
				},
			},
		},
	}
}

func (d *serverAddonsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverAddonsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverAddonsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.Get(fmt.Sprintf("/order/server_addon/%d/product", data.ServerNumber.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading server addons", err.Error())
		return
	}

	var apiResp []serverAddonAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing server addons response", err.Error())
		return
	}

	for _, item := range apiResp {
		a := item.Addon
		data.Addons = append(data.Addons, serverAddonModel{
			ID:    types.StringValue(a.ID),
			Name:  types.StringValue(a.Name),
			Type:  types.StringValue(a.Type),
			Price: apiPricesToModel(a.Price),
		})
	}

	if data.Addons == nil {
		data.Addons = []serverAddonModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
