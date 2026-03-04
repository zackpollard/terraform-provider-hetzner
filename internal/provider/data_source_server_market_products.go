// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zack/terraform-provider-hetzner/internal/client"
)

var _ datasource.DataSource = &serverMarketProductsDataSource{}

func NewServerMarketProductsDataSource() datasource.DataSource {
	return &serverMarketProductsDataSource{}
}

type serverMarketProductsDataSource struct {
	client *client.Client
}

type serverMarketProductsModel struct {
	Products []serverMarketProductModel `tfsdk:"products"`
}

type serverMarketProductModel struct {
	ID              types.Int64                   `tfsdk:"id"`
	Name            types.String                  `tfsdk:"name"`
	Description     []types.String                `tfsdk:"description"`
	Traffic         types.String                  `tfsdk:"traffic"`
	Dist            []types.String                `tfsdk:"dist"`
	CPU             types.String                  `tfsdk:"cpu"`
	CPUBenchmark    types.Int64                   `tfsdk:"cpu_benchmark"`
	MemorySize      types.Int64                   `tfsdk:"memory_size"`
	HDDSize         types.Int64                   `tfsdk:"hdd_size"`
	HDDText         types.String                  `tfsdk:"hdd_text"`
	HDDCount        types.Int64                   `tfsdk:"hdd_count"`
	Datacenter      types.String                  `tfsdk:"datacenter"`
	NetworkSpeed    types.String                  `tfsdk:"network_speed"`
	Price           types.String                  `tfsdk:"price"`
	PriceHourly     types.String                  `tfsdk:"price_hourly"`
	FixedPrice      types.Bool                    `tfsdk:"fixed_price"`
	NextReduceDate  types.String                  `tfsdk:"next_reduce_date"`
	OrderableAddons []serverOrderProductAddonItem `tfsdk:"orderable_addons"`
}

// API response types

type serverMarketProductAPIResponse struct {
	Product serverMarketProductAPI `json:"product"`
}

type serverMarketProductAPI struct {
	ID              int64                          `json:"id"`
	Name            string                         `json:"name"`
	Description     []string                       `json:"description"`
	Traffic         string                         `json:"traffic"`
	Dist            []string                       `json:"dist"`
	CPU             string                         `json:"cpu"`
	CPUBenchmark    int64                          `json:"cpu_benchmark"`
	MemorySize      int64                          `json:"memory_size"`
	HDDSize         int64                          `json:"hdd_size"`
	HDDText         string                         `json:"hdd_text"`
	HDDCount        int64                          `json:"hdd_count"`
	Datacenter      string                         `json:"datacenter"`
	NetworkSpeed    string                         `json:"network_speed"`
	Price           string                         `json:"price"`
	PriceHourly     string                         `json:"price_hourly"`
	FixedPrice      bool                           `json:"fixed_price"`
	NextReduceDate  string                         `json:"next_reduce_date"`
	OrderableAddons []serverOrderProductAddonAPI   `json:"orderable_addons"`
}

func (d *serverMarketProductsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_market_products"
}

func (d *serverMarketProductsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Hetzner Server Auction (server market) products.",
		Attributes: map[string]schema.Attribute{
			"products": schema.ListNestedAttribute{
				MarkdownDescription: "List of server market products.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Product ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Product name.",
							Computed:            true,
						},
						"description": schema.ListAttribute{
							MarkdownDescription: "Product description lines.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"traffic": schema.StringAttribute{
							MarkdownDescription: "Included traffic.",
							Computed:            true,
						},
						"dist": schema.ListAttribute{
							MarkdownDescription: "Available distributions.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"cpu": schema.StringAttribute{
							MarkdownDescription: "CPU model.",
							Computed:            true,
						},
						"cpu_benchmark": schema.Int64Attribute{
							MarkdownDescription: "CPU benchmark score.",
							Computed:            true,
						},
						"memory_size": schema.Int64Attribute{
							MarkdownDescription: "Memory size in GB.",
							Computed:            true,
						},
						"hdd_size": schema.Int64Attribute{
							MarkdownDescription: "HDD size in GB.",
							Computed:            true,
						},
						"hdd_text": schema.StringAttribute{
							MarkdownDescription: "HDD description text.",
							Computed:            true,
						},
						"hdd_count": schema.Int64Attribute{
							MarkdownDescription: "Number of HDDs.",
							Computed:            true,
						},
						"datacenter": schema.StringAttribute{
							MarkdownDescription: "Datacenter location.",
							Computed:            true,
						},
						"network_speed": schema.StringAttribute{
							MarkdownDescription: "Network speed.",
							Computed:            true,
						},
						"price": schema.StringAttribute{
							MarkdownDescription: "Monthly price.",
							Computed:            true,
						},
						"price_hourly": schema.StringAttribute{
							MarkdownDescription: "Hourly price.",
							Computed:            true,
						},
						"fixed_price": schema.BoolAttribute{
							MarkdownDescription: "Whether the price is fixed.",
							Computed:            true,
						},
						"next_reduce_date": schema.StringAttribute{
							MarkdownDescription: "Date of next price reduction.",
							Computed:            true,
						},
						"orderable_addons": schema.ListNestedAttribute{
							MarkdownDescription: "Available addons.",
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
									"min": schema.Int64Attribute{
										MarkdownDescription: "Minimum quantity.",
										Computed:            true,
									},
									"max": schema.Int64Attribute{
										MarkdownDescription: "Maximum quantity.",
										Computed:            true,
									},
									"prices": schema.ListNestedAttribute{
										MarkdownDescription: "Addon pricing per location.",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: priceSchemaAttrs,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *serverMarketProductsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverMarketProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/order/server_market/product")
	if err != nil {
		resp.Diagnostics.AddError("Error reading server market products", err.Error())
		return
	}

	var apiResp []serverMarketProductAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing server market products response", err.Error())
		return
	}

	var data serverMarketProductsModel
	for _, item := range apiResp {
		p := item.Product
		product := serverMarketProductModel{
			ID:             types.Int64Value(p.ID),
			Name:           types.StringValue(p.Name),
			Traffic:        types.StringValue(p.Traffic),
			CPU:            types.StringValue(p.CPU),
			CPUBenchmark:   types.Int64Value(p.CPUBenchmark),
			MemorySize:     types.Int64Value(p.MemorySize),
			HDDSize:        types.Int64Value(p.HDDSize),
			HDDText:        types.StringValue(p.HDDText),
			HDDCount:       types.Int64Value(p.HDDCount),
			Datacenter:     types.StringValue(p.Datacenter),
			NetworkSpeed:   types.StringValue(p.NetworkSpeed),
			Price:          types.StringValue(p.Price),
			PriceHourly:    types.StringValue(p.PriceHourly),
			FixedPrice:     types.BoolValue(p.FixedPrice),
			NextReduceDate: types.StringValue(p.NextReduceDate),
		}

		for _, d := range p.Description {
			product.Description = append(product.Description, types.StringValue(d))
		}
		if product.Description == nil {
			product.Description = []types.String{}
		}

		for _, d := range p.Dist {
			product.Dist = append(product.Dist, types.StringValue(d))
		}
		if product.Dist == nil {
			product.Dist = []types.String{}
		}

		product.OrderableAddons = apiAddonsToModel(p.OrderableAddons)

		data.Products = append(data.Products, product)
	}

	if data.Products == nil {
		data.Products = []serverMarketProductModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
