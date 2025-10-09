// Package models design for structs for requests
package models

type Price struct {
	PriceWithoutDiscount int `json:"basic"`
	PriceWithDiscount    int `json:"product"`
}

type Size struct {
	Price Price `json:"price"`
}

type Color struct {
	Name string `json:"name"`
}

type Product struct {
	Articul  int     `json:"id"`
	RcID     int     `json:"rcId"`
	Brand    string  `json:"brand"`
	Colors   []Color `json:"colors"`
	Name     string  `json:"name"`
	Entity   string  `json:"entity"`
	Supplier string  `json:"supplier"`
	Sizes    []Size  `json:"sizes"`
}

type ResponseBody struct {
	Products []Product `json:"products"`
}
