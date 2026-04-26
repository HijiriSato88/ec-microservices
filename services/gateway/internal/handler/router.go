package handler

import (
	"encoding/json"
	"net/http"

	cartpb "github.com/hijiri/ec-microservices/gen/go/cart"
	productpb "github.com/hijiri/ec-microservices/gen/go/product"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	product productpb.ProductServiceClient
	cart    cartpb.CartServiceClient
}

func New(product productpb.ProductServiceClient, cart cartpb.CartServiceClient) *Handler {
	return &Handler{product: product, cart: cart}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	h.registerProductRoutes(mux)
	h.registerCartRoutes(mux)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, _ := status.FromError(err)
	switch st.Code() {
	case codes.NotFound:
		http.Error(w, st.Message(), http.StatusNotFound)
	case codes.InvalidArgument:
		http.Error(w, st.Message(), http.StatusBadRequest)
	default:
		http.Error(w, st.Message(), http.StatusInternalServerError)
	}
}
