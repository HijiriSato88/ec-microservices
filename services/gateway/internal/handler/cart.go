package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	cartpb "github.com/hijiri/ec-microservices/gen/go/cart"
)

func (h *Handler) registerCartRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /cart/{user_id}", h.getCart)
	mux.HandleFunc("POST /cart/{user_id}/items", h.addItem)
	mux.HandleFunc("DELETE /cart/{user_id}/items/{product_id}", h.removeItem)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	resp, err := h.cart.GetCart(context.Background(), &cartpb.GetCartRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.Cart)
}

func (h *Handler) addItem(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	var body struct {
		ProductID int64 `json:"product_id"`
		Quantity  int32 `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	resp, err := h.cart.AddItem(context.Background(), &cartpb.AddItemRequest{
		UserId:    userID,
		ProductId: body.ProductID,
		Quantity:  body.Quantity,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.Cart)
}

func (h *Handler) removeItem(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	productID, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid product_id", http.StatusBadRequest)
		return
	}
	resp, err := h.cart.RemoveItem(context.Background(), &cartpb.RemoveItemRequest{
		UserId:    userID,
		ProductId: productID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.Cart)
}
