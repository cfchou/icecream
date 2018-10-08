package model

// Product is everything about an ice cream product.
type Product struct {
	// ProductId is mandatory and acts as the primary key.
	ProductID string `json:"productId"`
	Name      string `json:"name"`

	ImageClosed           string   `json:"image_closed"`
	ImageOpen             string   `json:"image_open"`
	Description           string   `json:"description"`
	Story                 string   `json:"story"`
	SourcingValues        []string `json:"sourcing_values"`
	Ingredients           []string `json:"ingredients"`
	AllergyInfo           string   `json:"allergy_info"`
	DietaryCertifications string   `json:"dietary_certifications"`
}

// Products is a page of Products.
type Products struct {
	// When Cursor is presented, it marks the last row of this page and could
	// be used to query the next page.
	Cursor string `json:"cursor,omitempty"`

	Products []Product `json:"products"`
}
