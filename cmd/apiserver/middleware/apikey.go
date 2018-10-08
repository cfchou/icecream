package handler

import "bitbucket.org/cfchou/icecream/pkg/backend/model"

type APIKeyBackend interface {
	Authenticate(apiKey string) error
}

type ProductBackend interface {
	// Create exclusively(must not existed)
	Create(product *model.Product) error

	Read(productID string) (*model.Product, error)
	// Don't provide ReadAll() because in real world there should be always a
	// limit for the amount to read.
	ReadMany(cursor string, limit int) (*model.Products, error)

	//Update existing
	Update(product *model.Product) error

	// TODO: partial update
	//UpdatePartial(productID string, kvs map[string]interface{}) error

	// Create or replace(fully update)
	Upsert(product *model.Product) error

	// Delete
	Delete(productID string) error
}
