package handler

import (
	"encoding/json"
	"fmt"
	"github.com/cfchou/icecream/pkg/backend/model"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

// ProductBackend is an interface for backends capable of accessing Product
type ProductBackend interface {
	// Create exclusively creates product. Success only if no Product with the
	// same ProductId existed.
	Create(product *model.Product) error

	// Read finds the Product with the given productID. Return error if not found.
	Read(productID string) (*model.Product, error)

	// ReadMany reads a page of products. Cursor is from last ReadMany and
	// represents the end of the previous page. Limit is the number of Products
	// that will be returned in a page. If cursor is empty then ReadMany begins
	// from the first page. Limit must be larger than 0. Return error if no
	// product read.
	ReadMany(cursor string, limit int) (*model.Products, error)

	// Update updates product. Success only if a Product with the same ProductId
	// existed. Return error if not existed.
	Update(product *model.Product) error

	// UpdatePartial updates product. Success only if a Product with the same
	// ProductId existed. Return error if not existed.
	UpdatePartial(productID string, kvs map[string]interface{}) error

	// Upsert inserts product. If a Product with the same productID existed
	// already, then a replacement is performed.
	Upsert(product *model.Product) error

	// Delete the Product with productID
	Delete(productID string) error
}

// ProductHandler provides http handlers for various methods.
type ProductHandler struct {
	log         log15.Logger
	limitToRead int
	backend     ProductBackend
}

// HandleGet reads a Product with the given productID retrieved
// from the url.
func (h *ProductHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID, ok := params["productID"]
	if !ok {
		// Bad Request
		w.WriteHeader(400)
		return
	}
	product, err := h.backend.Read(productID)
	if err != nil {
		// Not Found
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	bs, err := json.Marshal(product)
	if err != nil {
		// Internal Server Error
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

// HandleGetMany reads a page of products. Query parameters may include "cursor"
// and "limit". Cursor is from last HandleGetMany and represents the end of the
// previous page. Limit is the number of Products that will be returned in a
// page. If cursor is empty then ReadMany begins from the first page. Limit must
// be larger than 0.
func (h *ProductHandler) HandleGetMany(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	cursor := qs.Get("cursor")
	limit := qs.Get("limit")
	var limitToRead = h.limitToRead
	if limit != "" {
		n, err := strconv.Atoi(limit)
		if err != nil {
			// Bad Request
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		if n < 1 {
			// Bad Request
			w.WriteHeader(400)
			w.Write([]byte("Invalid limit"))
			return
		} else if n > limitToRead {
			h.log.Warn("limit exceeds limitToRead")
		} else {
			limitToRead = n
		}
	}
	mps, err := h.backend.ReadMany(cursor, limitToRead)
	if err != nil {
		// Not Found
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	bs, err := json.Marshal(mps)
	if err != nil {
		// Internal Server Error
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
	return
}

// HandlePost exclusively creates product. It unmarshals r.Body to
// model.Product. Note that every field in Product is required.
// It Success only if no Product with the same ProductId existed.
func (h *ProductHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Internal Server Error
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	// validate input
	product, err := validateProductFieldsAllPresented(body)
	if err != nil {
		// Bad Request
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	// productID must not be empty
	if product.ProductID == "" {
		// Bad Request
		w.WriteHeader(400)
		w.Write([]byte("invalid data"))
		return
	}
	// Create exclusively(product must not existed)
	if err := h.backend.Create(product); err != nil {
		// Forbidden
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	// Created
	w.WriteHeader(201)
	return
}

// HandlePut creates or replace(update fully) a product. If a Product with the
// same productID existed already, then a replacement is performed.
// Besides a productID retrieved from the url, it unmarshals r.Body to
// model.Product. Note that every field in Product is required.
func (h *ProductHandler) HandlePut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID, ok := params["productID"]
	if !ok {
		// Bad Request
		w.WriteHeader(400)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Internal Server Error
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	// validate input
	product, err := validateProductFieldsAllPresented(body)
	if err != nil {
		// Bad Request
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	if product.ProductID != "" && product.ProductID != productID {
		// Bad Request
		w.WriteHeader(400)
		w.Write([]byte("invalid data"))
		return
	}
	// sets product.productID if it's empty
	product.ProductID = productID

	// Upsert to ensure idempotent
	if err := h.backend.Upsert(product); err != nil {
		// Forbidden
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	// Created
	w.WriteHeader(201)
	return
}

func (h *ProductHandler) HandlePatch(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID, ok := params["productID"]
	if !ok {
		// Bad Request
		w.WriteHeader(400)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Internal Server Error
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal(body, &input); err != nil {
		// Bad Request
		w.WriteHeader(400)
		w.Write([]byte("invalid data"))
		return
	}

	if pid, ok := input["productId"]; ok && pid != productID {
		// Bad Request
		w.WriteHeader(400)
		w.Write([]byte("invalid data"))
		return
	}

	if err := h.backend.UpdatePartial(productID, input); err != nil {
		// Forbidden
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	return
}

// ProductHandle the Product with productID retrieved from the url.
func (h *ProductHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID, ok := params["productID"]
	if !ok {
		// Bad Request
		w.WriteHeader(400)
		return
	}
	if err := h.backend.Delete(productID); err != nil {
		// Forbidden
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
}

// Sanity check(inefficient)
// TODO: reflect to check data matching names and types of product fields.
func validateProductFieldsAllPresented(data []byte) (*model.Product, error) {
	var input map[string]interface{}
	var product model.Product
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// Check if input has only keys in keyMap
	keyMap := map[string]int8{
		"productId": 1, "name": 1, "image_closed": 1, "image_open": 1,
		"description": 1, "story": 1, "sourcing_values": 1, "ingredients": 1,
		"allergy_info": 1, "dietary_certifications": 1,
	}
	for k := range input {
		if _, ok := keyMap[k]; !ok {
			return nil, errors.New(fmt.Sprintf("extra field:%s", k))
		}
	}
	// Check if input missing any key
	for k := range keyMap {
		if _, ok := input[k]; !ok {
			return nil, errors.New(fmt.Sprintf("missing field:%s", k))
		}
	}
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

// CreateProductHandler creates ProductBackend with ProductBackend. Limit is the
// max number of products read in a page.
func CreateProductHandler(productBackend ProductBackend, limit int) *ProductHandler {
	return &ProductHandler{
		log:         log15.New("module", "handler.product"),
		limitToRead: limit,
		backend:     productBackend,
	}
}
