package mongodb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cfchou/icecream/pkg/backend/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/inconshreveable/log15"
	pe "github.com/pkg/errors"
)

const productsCollection = "products"

var (
	log = log15.New("module", "backend.mongodb")
	// ErrExisted when data existed in the db.
	ErrExisted = errors.New("existed")
	// ErrInconsistent when data violates some constraints, e.g. duplicated
	// ProductID
	ErrInconsistent = errors.New("inconsistent")
	// ErrParameters when inputs are invalid.
	ErrParameters = errors.New("bad parameters")
)

type mProduct struct {
	ID bson.ObjectId `bson:"_id,omitempty" json:"_id,omitempty"`
	// ProductId is mandatory.
	ProductID string `bson:"productId" json:"productId"`
	// Name is mandatory.
	Name string `bson:"name" json:"name"`

	ImageClosed           string   `bson:"image_closed" json:"image_closed"`
	ImageOpen             string   `bson:"image_open" json:"image_open"`
	Description           string   `bson:"description" json:"description"`
	Story                 string   `bson:"story" json:"story"`
	SourcingValues        []string `bson:"sourcing_values" json:"sourcing_values"`
	Ingredients           []string `bson:"ingredients" json:"ingredients"`
	AllergyInfo           string   `bson:"allergy_info" json:"allergy_info"`
	DietaryCertifications string   `bson:"dietary_certifications" json:"dietary_certifications"`
}

func createMProduct(product *model.Product) *mProduct {
	return &mProduct{
		ProductID:             product.ProductID,
		Name:                  product.Name,
		ImageClosed:           product.ImageClosed,
		ImageOpen:             product.ImageOpen,
		Description:           product.Description,
		Story:                 product.Story,
		SourcingValues:        product.SourcingValues,
		Ingredients:           product.Ingredients,
		AllergyInfo:           product.AllergyInfo,
		DietaryCertifications: product.DietaryCertifications,
	}
}

func (h *mProduct) ToProduct() *model.Product {
	return &model.Product{
		ProductID:             h.ProductID,
		Name:                  h.Name,
		ImageClosed:           h.ImageClosed,
		ImageOpen:             h.ImageOpen,
		Description:           h.Description,
		Story:                 h.Story,
		SourcingValues:        h.SourcingValues,
		Ingredients:           h.Ingredients,
		AllergyInfo:           h.AllergyInfo,
		DietaryCertifications: h.DietaryCertifications,
	}
}

// MongoProductBackend stores a mongoDB session to support CRUD for Product.
type MongoProductBackend struct {
	session *mgo.Session
}

// Close closes the internal mongoDB session.
func (h *MongoProductBackend) Close() {
	h.session.Close()
}

// Create exclusively creates product. Success only if no Product with the
// same ProductId existed.
func (h *MongoProductBackend) Create(product *model.Product) error {
	if product.ProductID == "" {
		log.Error("Invalid productId", "productId", product.ProductID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}
	mp := createMProduct(product)

	info, err := h.session.DB("").C(productsCollection).Upsert(
		&bson.M{"productId": mp.ProductID}, &bson.M{"$setOnInsert": mp})
	if err != nil {
		log.Error("Create failed", "productId", mp.ProductID, "err", err)
		return pe.WithStack(err)
	}
	// One matched but no-op since $set is not given.
	if info.Matched != 0 {
		log.Error("Create existed failed ", "productId", mp.ProductID,
			"err", ErrExisted)
		return pe.WithStack(ErrExisted)
	}
	// No one matched, trigger insert with $setOnInsert
	oid := info.UpsertedId.(bson.ObjectId)
	log.Debug(fmt.Sprintf("Create _id=%s", oid.Hex()),
		"productId", mp.ProductID)
	return nil
}

// Upsert inserts product. If a Product with the same productID existed already,
// then a replacement is performed.
func (h *MongoProductBackend) Upsert(product *model.Product) error {
	if product.ProductID == "" {
		log.Error("Invalid productId", "productId", product.ProductID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}
	mp := createMProduct(product)

	info, err := h.session.DB("").C(productsCollection).Upsert(
		&bson.M{"productId": mp.ProductID}, mp)
	if err != nil {
		log.Error("Upsert failed", "productId", mp.ProductID, "err", err)
		return pe.WithStack(err)
	}
	// One matched, trigger update
	if info.Updated != 0 {
		log.Debug("Upsert update succeeded", "productId", mp.ProductID)
		return nil
	}
	// No one matched, trigger insert
	oid := info.UpsertedId.(bson.ObjectId)
	log.Debug(fmt.Sprintf("Upsert insert _id=%s", oid.Hex()),
		"productId", mp.ProductID)
	return nil
}

// Update updates product. Success only if a Product with the same
// ProductId existed. Return error if not existed.
func (h *MongoProductBackend) Update(product *model.Product) error {
	if product.ProductID == "" {
		log.Error("Invalid productId", "productId", product.ProductID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}
	mp := createMProduct(product)

	if err := h.session.DB("").C(productsCollection).Update(
		&bson.M{"productId": mp.ProductID}, mp); err != nil {
		log.Error("Update failed ", "productId", mp.ProductID,
			"err", err)
		return pe.WithStack(err)
	}
	log.Debug("Update succeeded", "productId", mp.ProductID)
	return nil
}

// UpdatePartial updates product. Success only if a Product with the same
// ProductId existed. Return error if not existed.
func (h *MongoProductBackend) UpdatePartial(productID string, kvs map[string]interface{}) error {
	if productID == "" {
		log.Error("Invalid productId", "productId", productID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}

	if pid, ok := kvs["productId"]; ok && pid != productID {
		log.Error("Different productId", "productId", productID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}
	// Sanity check(inefficient)
	// TODO: reflect to check kvs matching names and types of product fields.

	// Check if input has only keys in keyMap
	keyMap := map[string]int8{
		"productId": 1, "name": 1, "image_closed": 1, "image_open": 1,
		"description": 1, "story": 1, "sourcing_values": 1, "ingredients": 1,
		"allergy_info": 1, "dietary_certifications": 1,
	}
	for k := range kvs {
		if _, ok := keyMap[k]; !ok {
			log.Error("Unknown extra field", "productId", productID,
				"err", ErrParameters)
			return pe.WithStack(ErrParameters)
		}
	}

	// Check if values of kvs fits types of fields of Product
	bs, err := json.Marshal(kvs)
	if err != nil {
		log.Error("json.Marshal failed ", "productId", productID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}

	var product model.Product
	if err := json.Unmarshal(bs, &product); err != nil {
		log.Error("json.Unmarshal failed ", "productId", productID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}

	// Safe to update
	if err := h.session.DB("").C(productsCollection).Update(
		&bson.M{"productId": productID}, bson.M{"$set": kvs}); err != nil {
		log.Error("Update failed ", "productId", productID,
			"err", err)
		return pe.WithStack(err)
	}
	log.Debug("Update succeeded", "productId", productID)
	return nil
}

// Read finds the Product with the given productID. Return error if not found.
func (h *MongoProductBackend) Read(productID string) (*model.Product, error) {
	if productID == "" {
		log.Error("Invalid productId", "productId", productID,
			"err", ErrParameters)
		return nil, pe.WithStack(ErrParameters)
	}
	var mps = make([]mProduct, 0)
	q := h.session.DB("").C(productsCollection).Find(
		&bson.M{"productId": productID})
	if err := q.All(&mps); err != nil {
		log.Error("Query.All failed", "productId", productID, "err", err)
		return nil, pe.WithStack(err)
	}
	if len(mps) == 0 {
		return nil, pe.WithStack(mgo.ErrNotFound)
	} else if len(mps) > 1 {
		// By design this should not happen. Most likely a duplicated product is
		// added in an out-of-band fashion.
		log.Error("Find gets more than 1", "productId", productID,
			"err", ErrInconsistent)
		return nil, pe.WithStack(ErrInconsistent)
	}
	log.Debug(fmt.Sprintf("Find _id=%s", mps[0].ID.Hex()),
		"productId", mps[0].ProductID)
	return mps[0].ToProduct(), nil
}

// ReadMany reads a page of products. Cursor is from last ReadMany and
// represents the end of the previous page. Limit is the number of Products that
// will be returned in a page. If cursor is empty then ReadMany begins from the
// first page. Limit must be larger than 0. Return error if no product read.
func (h *MongoProductBackend) ReadMany(cursor string, limit int) (*model.Products, error) {
	if limit <= 0 {
		log.Error(fmt.Sprintf("Invalid limit:%d", limit), "err", ErrParameters)
		return nil, pe.WithStack(ErrParameters)
	}
	selector := &bson.M{}
	if cursor != "" && bson.IsObjectIdHex(cursor) {
		objectID := bson.ObjectIdHex(cursor)
		selector = &bson.M{
			"_id": &bson.M{"$gt": objectID},
		}
	}
	var mps []mProduct
	q := h.session.DB("").C(productsCollection).Find(selector).Limit(limit)
	if err := q.All(&mps); err != nil {
		log.Error("Query.All failed", "from", cursor, "err", err)
		return nil, pe.WithStack(err)
	}
	if len(mps) == 0 {
		return nil, pe.WithStack(mgo.ErrNotFound)
	}
	ret := &model.Products{
		Cursor:   mps[len(mps)-1].ID.Hex(),
		Products: make([]model.Product, 0),
	}
	for _, mp := range mps {
		ret.Products = append(ret.Products, *mp.ToProduct())
	}
	log.Debug("ReadMany succeeded", "count", len(mps), "from", cursor,
		"to", ret.Cursor)
	return ret, nil
}

// Delete the Product with productID
func (h *MongoProductBackend) Delete(productID string) error {
	if productID == "" {
		log.Error("Invalid productId", "productId", productID,
			"err", ErrParameters)
		return pe.WithStack(ErrParameters)
	}
	if err := h.session.DB("").C(productsCollection).Remove(&bson.M{
		"productId": productID,
	}); err != nil {
		log.Error("Remove failed", "productId", productID, "err", err)
		return pe.WithStack(err)
	}
	log.Debug("Remove succeeded", "productId", productID)
	return nil
}

// CreateMongoProductBackend creates MongoProductBackend
func CreateMongoProductBackend(session *mgo.Session) (*MongoProductBackend, error) {
	return &MongoProductBackend{
		session: session,
	}, nil
}
