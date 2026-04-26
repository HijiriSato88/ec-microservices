package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	productpb "github.com/hijiri/ec-microservices/gen/go/product"
)

func (h *Handler) registerProductRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /products", h.listProducts)
	mux.HandleFunc("GET /products/{id}", h.getProduct)
	mux.HandleFunc("POST /products", h.createProduct)
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	resp, err := h.product.ListProducts(context.Background(), &productpb.ListProductsRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.Products)
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	resp, err := h.product.GetProduct(context.Background(), &productpb.GetProductRequest{Id: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.Product)
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req productpb.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	resp, err := h.product.CreateProduct(context.Background(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.Product)
}
