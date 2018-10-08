package model

// Product is everything about an ice cream product.
type Product struct {
	// ProductId is mandatory.
	ProductID string `json:"productId"`
	// Name is mandatory.
	Name string `json:"name"`

	ImageClosed           string   `json:"image_closed"`
	ImageOpen             string   `json:"image_open"`
	Description           string   `json:"description"`
	Story                 string   `json:"story"`
	SourcingValues        []string `json:"sourcing_values"`
	Ingredients           []string `json:"ingredients"`
	AllergyInfo           string   `json:"allergy_info"`
	DietaryCertifications string   `json:"dietary_certifications"`
}

// Products support pagination when reading a large chunk of Products.
type Products struct {
	// When Cursor is presented, client can use it as the start to query the
	// next page.
	Cursor string `json:"cursor,omitempty"`

	Products []Product `json:"products"`
}

