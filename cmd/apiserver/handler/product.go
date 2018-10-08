package handler

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type ProductHandler struct {
	APIKeyBackend  APIKeyBackend
	ProductBackend ProductBackend
}

func (h *ProductHandler) HandleGet(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func (h *ProductHandler) HandlePost(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func (h *ProductHandler) HandlePut(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func (h *ProductHandler) HandleDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
}

func CreateProductHandler(apiKeyBackend APIKeyBackend, productBackend ProductBackend) *ProductHandler {
	return &ProductHandler{
		APIKeyBackend: apiKeyBackend,
		ProductBackend: productBackend,
	}
}
