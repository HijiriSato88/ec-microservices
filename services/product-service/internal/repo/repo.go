package repo

import pb "github.com/hijiri/ec-microservices/gen/go/product"

// Repository は product-service のデータアクセス層のインターフェース。
// in-memory と MySQL で同じインターフェースを実装する。
type Repository interface {
	FindByID(id int64) (*pb.Product, error)
	FindAll() ([]*pb.Product, error)
	Save(name, description string, price int64) (*pb.Product, error)
}
