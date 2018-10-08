package mongodb

import (
	"bitbucket.org/cfchou/icecream/pkg/backend/model"
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/inconshreveable/log15"
	pe "github.com/pkg/errors"
)

var (
	log             = log15.New("module", "backend.mongodb")
	ErrExisted      = errors.New("existed")
	ErrInconsistent = errors.New("inconsistent")
	ErrParameters   = errors.New("bad parameters")
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

type MongoProductBackend struct {
	session *mgo.Session
}

// Close close the mongoDB session.
func (h *MongoProductBackend) Close() {
	h.session.Close()
}

// Create exclusively creates a product. Success only if no Product with the
// same ProductId existed.
func (h *MongoProductBackend) Create(product *model.Product) error {
	mp := createMProduct(product)

	info, err := h.session.DB("").C("products").Upsert(
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

/*
TODO: Partial update
func (h *MongoProductBackend) UpdatePartial(productID string, kvs map[string]interface{}) error {
	// TODO: reflect or hard-code to check kvs matching names and types of
	// product fileds.
	err := h.session.DB("").C("products").Update(
		&bson.M{"productId": productID}, bson.M{ "$set": kvs})
}
*/

// Update updates a product. Success only if a Product with the same
// ProductId existed.
func (h *MongoProductBackend) Update(product *model.Product) error {
	mp := createMProduct(product)

	err := h.session.DB("").C("products").Update(
		&bson.M{"productId": mp.ProductID}, mp)
	if err != nil {
		log.Error("Update failed ", "productId", mp.ProductID,
			"err", err)
		return pe.WithStack(err)
	}
	log.Debug("Update succeeded", "productId", mp.ProductID)
	return nil
}

func (h *MongoProductBackend) Upsert(product *model.Product) error {
	mp := createMProduct(product)

	info, err := h.session.DB("").C("products").Upsert(
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

func (h *MongoProductBackend) Read(productID string) (*model.Product, error) {
	var mps []mProduct
	q := h.session.DB("").C("products").Find(
		&bson.M{"productId": productID})
	if err := q.All(&mps); err != nil {
		log.Error("Query.All failed", "productId", productID, "err", err)
		return nil, pe.WithStack(err)
	}
	if len(mps) == 0 {
		return nil, pe.WithStack(mgo.ErrNotFound)
	} else if len(mps) != 1 {
		// By design this should not happen. Most likely a duplicated product is
		// added in an out-of-band fashion.
		log.Error("Read gets more than 1", "productId", productID,
			"err", ErrInconsistent)
		return nil, pe.WithStack(ErrInconsistent)
	}
	log.Debug(fmt.Sprintf("Read _id=%s", mps[0].ID.Hex()),
		"productId", mps[0].ProductID)
	return mps[0].ToProduct(), nil
}

func (h *MongoProductBackend) ReadMany(cursor string, limit int) (*model.Products, error) {
	if limit <= 0 {
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
	q := h.session.DB("").C("products").Find(selector).Limit(limit)
	if err := q.All(&mps); err != nil {
		log.Error("Query.All failed", "from", cursor, "err", err)
		return nil, pe.WithStack(err)
	}
	if len(mps) == 0 {
		/*
			log.Debug("ReadMany succeeded", "count", 0, "from", cursor)
			return &model.Products{
				Products: make([]model.Product, 0),
			}, nil
		*/
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

func (h *MongoProductBackend) Delete(productID string) error {
	if err := h.session.DB("").C("products").Remove(&bson.M{
		"productId": productID,
	}); err != nil {
		log.Error("Remove failed", "productId", productID, "err", err)
		return pe.WithStack(err)
	}
	log.Debug("Remove succeeded", "productId", productID)
	return nil
}

func (h *MongoProductBackend) Authenticate(apiKey string) error {
	return nil
}

func CreateMongoProductBackend(url string) (*MongoProductBackend, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		log.Error("mgo.Dial failed", "err", err.Error())
		return nil, pe.WithStack(err)
	}
	return &MongoProductBackend{
		session: session,
	}, nil
}
