package main

import (
	"log"
	"net/http"
	"os"

	cartpb "github.com/hijiri/ec-microservices/gen/go/cart"
	productpb "github.com/hijiri/ec-microservices/gen/go/product"
	"github.com/hijiri/ec-microservices/services/gateway/internal/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	productAddr := os.Getenv("PRODUCT_ADDR")
	if productAddr == "" {
		productAddr = "localhost:50051"
	}

	cartAddr := os.Getenv("CART_ADDR")
	if cartAddr == "" {
		cartAddr = "localhost:50052"
	}

	productConn, err := grpc.NewClient(productAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to product-service: %v", err)
	}
	defer productConn.Close()

	cartConn, err := grpc.NewClient(cartAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to cart-service: %v", err)
	}
	defer cartConn.Close()

	h := handler.New(
		productpb.NewProductServiceClient(productConn),
		cartpb.NewCartServiceClient(cartConn),
	)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	log.Printf("gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
