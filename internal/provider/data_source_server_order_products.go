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

var _ datasource.DataSource = &serverOrderProductsDataSource{}

func NewServerOrderProductsDataSource() datasource.DataSource {
	return &serverOrderProductsDataSource{}
}

type serverOrderProductsDataSource struct {
	client *client.Client
}

type serverOrderProductsModel struct {
	Products []serverOrderProductModel `tfsdk:"products"`
}

type serverOrderProductModel struct {
	ID              types.String                  `tfsdk:"id"`
	Name            types.String                  `tfsdk:"name"`
	Description     []types.String                `tfsdk:"description"`
	Traffic         types.String                  `tfsdk:"traffic"`
	Dist            []types.String                `tfsdk:"dist"`
	Arch            []types.Int64                 `tfsdk:"arch"`
	Lang            []types.String                `tfsdk:"lang"`
	Location        []types.String                `tfsdk:"location"`
	Prices          []serverOrderProductPrice     `tfsdk:"prices"`
	OrderableAddons []serverOrderProductAddonItem `tfsdk:"orderable_addons"`
}

type serverOrderProductPrice struct {
	Location        types.String `tfsdk:"location"`
	PriceNet        types.String `tfsdk:"price_net"`
	PriceGross      types.String `tfsdk:"price_gross"`
	SetupPriceNet   types.String `tfsdk:"setup_price_net"`
	SetupPriceGross types.String `tfsdk:"setup_price_gross"`
}

type serverOrderProductAddonItem struct {
	ID     types.String              `tfsdk:"id"`
	Name   types.String              `tfsdk:"name"`
	Min    types.Int64               `tfsdk:"min"`
	Max    types.Int64               `tfsdk:"max"`
	Prices []serverOrderProductPrice `tfsdk:"prices"`
}

// API response types

type serverOrderProductAPIResponse struct {
	Product serverOrderProductAPI `json:"product"`
}

type serverOrderProductAPI struct {
	ID              string                       `json:"id"`
	Name            string                       `json:"name"`
	Description     []string                     `json:"description"`
	Traffic         string                       `json:"traffic"`
	Dist            []string                     `json:"dist"`
	Arch            []int64                      `json:"arch"`
	Lang            []string                     `json:"lang"`
	Location        []string                     `json:"location"`
	Prices          []serverOrderProductPriceAPI `json:"prices"`
	OrderableAddons []serverOrderProductAddonAPI `json:"orderable_addons"`
}

type priceAmountAPI struct {
	Net   string `json:"net"`
	Gross string `json:"gross"`
}

type serverOrderProductPriceAPI struct {
	Location   string         `json:"location"`
	Price      priceAmountAPI `json:"price"`
	PriceSetup priceAmountAPI `json:"price_setup"`
}

type serverOrderProductAddonAPI struct {
	ID     string                       `json:"id"`
	Name   string                       `json:"name"`
	Min    int64                        `json:"min"`
	Max    int64                        `json:"max"`
	Prices []serverOrderProductPriceAPI `json:"prices"`
}

func (d *serverOrderProductsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_order_products"
}

var priceSchemaAttrs = map[string]schema.Attribute{
	"location": schema.StringAttribute{
		MarkdownDescription: "Location code.",
		Computed:            true,
	},
	"price_net": schema.StringAttribute{
		MarkdownDescription: "Net price.",
		Computed:            true,
	},
	"price_gross": schema.StringAttribute{
		MarkdownDescription: "Gross price.",
		Computed:            true,
	},
	"setup_price_net": schema.StringAttribute{
		MarkdownDescription: "Net setup price.",
		Computed:            true,
	},
	"setup_price_gross": schema.StringAttribute{
		MarkdownDescription: "Gross setup price.",
		Computed:            true,
	},
}

func (d *serverOrderProductsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available Hetzner dedicated server order products.",
		Attributes: map[string]schema.Attribute{
			"products": schema.ListNestedAttribute{
				MarkdownDescription: "List of server products.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
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
						"arch": schema.ListAttribute{
							MarkdownDescription: "Supported architectures.",
							Computed:            true,
							ElementType:         types.Int64Type,
						},
						"lang": schema.ListAttribute{
							MarkdownDescription: "Available languages.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"location": schema.ListAttribute{
							MarkdownDescription: "Available locations.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"prices": schema.ListNestedAttribute{
							MarkdownDescription: "Pricing information per location.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: priceSchemaAttrs,
							},
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

func (d *serverOrderProductsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverOrderProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	body, err := d.client.Get("/order/server/product")
	if err != nil {
		resp.Diagnostics.AddError("Error reading server order products", err.Error())
		return
	}

	var apiResp []serverOrderProductAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing server order products response", err.Error())
		return
	}

	var data serverOrderProductsModel
	for _, item := range apiResp {
		p := item.Product
		product := serverOrderProductModel{
			ID:      types.StringValue(p.ID),
			Name:    types.StringValue(p.Name),
			Traffic: types.StringValue(p.Traffic),
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

		for _, a := range p.Arch {
			product.Arch = append(product.Arch, types.Int64Value(a))
		}
		if product.Arch == nil {
			product.Arch = []types.Int64{}
		}

		for _, l := range p.Lang {
			product.Lang = append(product.Lang, types.StringValue(l))
		}
		if product.Lang == nil {
			product.Lang = []types.String{}
		}

		for _, loc := range p.Location {
			product.Location = append(product.Location, types.StringValue(loc))
		}
		if product.Location == nil {
			product.Location = []types.String{}
		}

		product.Prices = apiPricesToModel(p.Prices)
		product.OrderableAddons = apiAddonsToModel(p.OrderableAddons)

		data.Products = append(data.Products, product)
	}

	if data.Products == nil {
		data.Products = []serverOrderProductModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func apiPricesToModel(prices []serverOrderProductPriceAPI) []serverOrderProductPrice {
	result := make([]serverOrderProductPrice, 0, len(prices))
	for _, p := range prices {
		result = append(result, serverOrderProductPrice{
			Location:        types.StringValue(p.Location),
			PriceNet:        types.StringValue(p.Price.Net),
			PriceGross:      types.StringValue(p.Price.Gross),
			SetupPriceNet:   types.StringValue(p.PriceSetup.Net),
			SetupPriceGross: types.StringValue(p.PriceSetup.Gross),
		})
	}
	return result
}

func apiAddonsToModel(addons []serverOrderProductAddonAPI) []serverOrderProductAddonItem {
	result := make([]serverOrderProductAddonItem, 0, len(addons))
	for _, a := range addons {
		result = append(result, serverOrderProductAddonItem{
			ID:     types.StringValue(a.ID),
			Name:   types.StringValue(a.Name),
			Min:    types.Int64Value(a.Min),
			Max:    types.Int64Value(a.Max),
			Prices: apiPricesToModel(a.Prices),
		})
	}
	return result
}
