package main

import (
	"log"
	"net"
	"os"

	pb "github.com/hijiri/ec-microservices/gen/go/product"
	"github.com/hijiri/ec-microservices/services/product-service/internal"
	"github.com/hijiri/ec-microservices/services/product-service/internal/repo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	r := repo.NewMemoryRepo()
	srv := grpc.NewServer()
	pb.RegisterProductServiceServer(srv, internal.NewProductServer(r))
	reflection.Register(srv)

	log.Printf("product-service listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
