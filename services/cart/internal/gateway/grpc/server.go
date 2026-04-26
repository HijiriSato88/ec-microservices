package grpc

import (
	"context"

	pb "github.com/hijiri/ec-microservices/gen/go/cart"
	"github.com/hijiri/ec-microservices/services/cart/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartServer struct {
	pb.UnimplementedCartServiceServer
	repo          repository.Repository
	productClient *ProductClient
}

func NewCartServer(r repository.Repository, pc *ProductClient) *CartServer {
	return &CartServer{repo: r, productClient: pc}
}

func (s *CartServer) GetCart(_ context.Context, req *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	cart, err := s.repo.FindByUserID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}
	return &pb.GetCartResponse{Cart: toProto(cart)}, nil
}

func (s *CartServer) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	// product-service から商品情報を取得
	product, err := s.productClient.GetProduct(ctx, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product %d not found: %v", req.ProductId, err)
	}

	cart, err := s.repo.FindByUserID(req.UserId)
	if err != nil {
		cart = &repository.Cart{UserID: req.UserId}
	}

	// 既存アイテムがあれば数量を加算
	for _, item := range cart.Items {
		if item.ProductID == req.ProductId {
			item.Quantity += req.Quantity
			if err := s.repo.Save(cart); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to save cart: %v", err)
			}
			return &pb.AddItemResponse{Cart: toProto(cart)}, nil
		}
	}

	// 新規アイテムを追加
	cart.Items = append(cart.Items, &repository.CartItem{
		ProductID:   product.Id,
		ProductName: product.Name,
		Price:       product.Price,
		Quantity:    req.Quantity,
	})

	if err := s.repo.Save(cart); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save cart: %v", err)
	}
	return &pb.AddItemResponse{Cart: toProto(cart)}, nil
}

func (s *CartServer) RemoveItem(_ context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	cart, err := s.repo.FindByUserID(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cart not found: %v", err)
	}

	items := make([]*repository.CartItem, 0)
	for _, item := range cart.Items {
		if item.ProductID != req.ProductId {
			items = append(items, item)
		}
	}
	cart.Items = items

	if err := s.repo.Save(cart); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save cart: %v", err)
	}
	return &pb.RemoveItemResponse{Cart: toProto(cart)}, nil
}

func toProto(cart *repository.Cart) *pb.Cart {
	if cart == nil {
		return &pb.Cart{}
	}
	var total int64
	items := make([]*pb.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, &pb.CartItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
		total += item.Price * int64(item.Quantity)
	}
	return &pb.Cart{
		UserId: cart.UserID,
		Items:  items,
		Total:  total,
	}
}
