package internal

import (
	"context"

	pb "github.com/hijiri/ec-microservices/gen/go/product"
	"github.com/hijiri/ec-microservices/services/product-service/internal/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServer struct {
	pb.UnimplementedProductServiceServer
	repo repo.Repository
}

func NewProductServer(r repo.Repository) *ProductServer {
	return &ProductServer{repo: r}
}

func (s *ProductServer) GetProduct(_ context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.repo.FindByID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product %d not found", req.Id)
	}
	return &pb.GetProductResponse{Product: p}, nil
}

func (s *ProductServer) ListProducts(_ context.Context, _ *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := s.repo.FindAll()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}
	return &pb.ListProductsResponse{Products: products}, nil
}

func (s *ProductServer) CreateProduct(_ context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	p, err := s.repo.Save(req.Name, req.Description, req.Price)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}
	return &pb.CreateProductResponse{Product: p}, nil
}
