package repository

import pb "github.com/hijiri/ec-microservices/gen/go/product"

type Repository interface {
	FindByID(id int64) (*pb.Product, error)
	FindAll() ([]*pb.Product, error)
	Save(name, description string, price int64) (*pb.Product, error)
}
