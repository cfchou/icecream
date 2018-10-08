package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cfchou/icecream/cmd/apiserver/mocks"
	"github.com/cfchou/icecream/pkg/backend/model"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductHandler_HandleGet(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"
	product := &model.Product{ProductID: productID}
	mB.On("Read", productID).Return(product, nil)

	ph := CreateProductHandler(mB, 10)
	r := mux.NewRouter()
	r.Methods("GET").Path("/products/{productID}").HandlerFunc(ph.HandleGet)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/products/"+productID, nil)
	r.ServeHTTP(writer, request)
	assert.Equal(t, 200, writer.Code)

	var result model.Product
	bs, _ := ioutil.ReadAll(writer.Body)
	json.Unmarshal(bs, &result)
	assert.EqualValues(t, *product, result)
}

func TestProductHandler_HandleGet_404CausedByBackend(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"

	mB.On("Read", productID).
		Return(nil, fmt.Errorf("any error"))
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("GET").Path("/products/{productID}").HandlerFunc(ph.HandleGet)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/products/"+productID, nil)
	r.ServeHTTP(writer, request)
	assert.Equal(t, 404, writer.Code)
}

func TestProductHandler_HandleGetMany(t *testing.T) {
	mB := &mocks.ProductBackend{}

	products := &model.Products{
		Products: []model.Product{
			{ProductID: "001"},
			{ProductID: "002"},
		},
	}

	mB.On("ReadMany", "", 10).Return(products, nil)
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("GET").Path("/products/").HandlerFunc(ph.HandleGetMany)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/products/", nil)
	r.ServeHTTP(writer, request)
	assert.Equal(t, 200, writer.Code)

	var result model.Products
	bs, _ := ioutil.ReadAll(writer.Body)
	json.Unmarshal(bs, &result)
	assert.EqualValues(t, *products, result)
}

func TestProductHandler_HandleGetMany_404CausedByBackend(t *testing.T) {
	mB := &mocks.ProductBackend{}

	mB.On("ReadMany", "", 10).
		Return(nil, fmt.Errorf("any error"))
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("GET").Path("/products/").HandlerFunc(ph.HandleGetMany)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/products/", nil)
	r.ServeHTTP(writer, request)
	assert.Equal(t, 404, writer.Code)
}

func TestProductHandler_HandlePost(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"
	product := &model.Product{ProductID: productID}

	mB.On("Create", mock.Anything).Return(nil)
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("POST").Path("/products/").HandlerFunc(ph.HandlePost)

	writer := httptest.NewRecorder()
	bs, _ := json.Marshal(product)
	request, _ := http.NewRequest("POST", "/products/", bytes.NewBuffer(bs))
	r.ServeHTTP(writer, request)
	assert.Equal(t, 201, writer.Code)
}

func TestProductHandler_HandlePost_403CausedByBackend(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"
	product := &model.Product{ProductID: productID}

	mB.On("Create", mock.Anything).Return(fmt.Errorf("any error"))
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("POST").Path("/products/").HandlerFunc(ph.HandlePost)

	writer := httptest.NewRecorder()
	bs, _ := json.Marshal(product)
	request, _ := http.NewRequest("POST", "/products/", bytes.NewBuffer(bs))
	r.ServeHTTP(writer, request)
	assert.Equal(t, 403, writer.Code)
}

func TestProductHandler_HandlePut(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"
	product := &model.Product{ProductID: productID}

	mB.On("Upsert", mock.Anything).Return(nil)
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("PUT").Path("/products/{productID}").HandlerFunc(ph.HandlePut)

	writer := httptest.NewRecorder()
	bs, _ := json.Marshal(product)
	request, _ := http.NewRequest("PUT", "/products/"+productID, bytes.NewBuffer(bs))
	r.ServeHTTP(writer, request)
	assert.Equal(t, 201, writer.Code)
}

func TestProductHandler_HandlePut_403CausedByBackend(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"
	product := &model.Product{ProductID: productID}

	mB.On("Upsert", mock.Anything).Return(fmt.Errorf("any error"))
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("PUT").Path("/products/{productID}").HandlerFunc(ph.HandlePut)

	writer := httptest.NewRecorder()
	bs, _ := json.Marshal(product)
	request, _ := http.NewRequest("PUT", "/products/"+productID, bytes.NewBuffer(bs))
	r.ServeHTTP(writer, request)
	assert.Equal(t, 403, writer.Code)
}

func TestProductHandler_HandleDelete(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"

	mB.On("Delete", productID).Return(nil)
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("DELETE").Path("/products/{productID}").HandlerFunc(ph.HandleDelete)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/products/"+productID, nil)
	r.ServeHTTP(writer, request)
	assert.Equal(t, 200, writer.Code)
}

func TestProductHandler_HandleDelete_403CausedByBackend(t *testing.T) {
	mB := &mocks.ProductBackend{}

	productID := "001"

	mB.On("Delete", productID).Return(fmt.Errorf("any error"))
	ph := CreateProductHandler(mB, 10)

	r := mux.NewRouter()
	r.Methods("DELETE").Path("/products/{productID}").HandlerFunc(ph.HandleDelete)

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/products/"+productID, nil)
	r.ServeHTTP(writer, request)
	assert.Equal(t, 403, writer.Code)
}

func TestCreateProductHandler(t *testing.T) {
	//y := []byte(`{"description":"","ingredients":["abc"],"productId":"001","sourcing_values":[]}`)
	x := map[string]interface{}{
		"productId":       "001",
		"description":     "whatever",
		"ingredients":     []string{"abc"},
		"sourcing_values": []string{},
		"cool":            123,
	}
	z := []byte(`{
    "allergy_info": "",
    "description": "",
    "dietary_certifications": "",
    "image_closed": "",
    "image_open": "",
    "ingredients": [],
    "name": "Test2",
    "productId": "001",
    "sourcing_values": [],
    "story": ""
}
`)
	bs, err := json.Marshal(x)
	if err != nil {
		t.Error(err.Error())
		return
	}
	bsstr := string(bs)
	t.Log(bsstr)
	var p model.Product
	if err := json.Unmarshal(bs, &p); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("done")
	var q model.Product
	if err := json.Unmarshal(z, &q); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("done")
}
