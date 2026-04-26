package grpc

import (
	"context"
	"fmt"

	pb "github.com/hijiri/ec-microservices/gen/go/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	client pb.ProductServiceClient
}

func NewProductClient(addr string) (*ProductClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product-service: %w", err)
	}
	return &ProductClient{client: pb.NewProductServiceClient(conn)}, nil
}

func (c *ProductClient) GetProduct(ctx context.Context, id int64) (*pb.Product, error) {
	resp, err := c.client.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}
