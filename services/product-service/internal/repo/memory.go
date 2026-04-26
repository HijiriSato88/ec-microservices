package repo

import (
	"fmt"
	"sync"

	pb "github.com/hijiri/ec-microservices/gen/go/product"
)

type memoryRepo struct {
	mu       sync.RWMutex
	products map[int64]*pb.Product
	nextID   int64
}

func NewMemoryRepo() Repository {
	return &memoryRepo{
		products: make(map[int64]*pb.Product),
		nextID:   1,
	}
}

func (r *memoryRepo) FindByID(id int64) (*pb.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.products[id]
	if !ok {
		return nil, fmt.Errorf("product %d not found", id)
	}
	return p, nil
}

func (r *memoryRepo) FindAll() ([]*pb.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]*pb.Product, 0, len(r.products))
	for _, p := range r.products {
		products = append(products, p)
	}
	return products, nil
}

func (r *memoryRepo) Save(name, description string, price int64) (*pb.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := &pb.Product{
		Id:          r.nextID,
		Name:        name,
		Description: description,
		Price:       price,
	}
	r.products[r.nextID] = p
	r.nextID++
	return p, nil
}
