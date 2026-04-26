package mysql

import (
	"database/sql"
	"fmt"

	pb "github.com/hijiri/ec-microservices/gen/go/product"
	"github.com/hijiri/ec-microservices/services/product/internal/repository"
)

type mysqlRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) repository.Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) FindByID(id int64) (*pb.Product, error) {
	row := r.db.QueryRow("SELECT id, name, description, price FROM products WHERE id = ?", id)

	p := &pb.Product{}
	if err := row.Scan(&p.Id, &p.Name, &p.Description, &p.Price); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product %d not found", id)
		}
		return nil, fmt.Errorf("failed to scan: %w", err)
	}
	return p, nil
}

func (r *mysqlRepo) FindAll() ([]*pb.Product, error) {
	rows, err := r.db.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	var products []*pb.Product
	for rows.Next() {
		p := &pb.Product{}
		if err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Price); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *mysqlRepo) Save(name, description string, price int64) (*pb.Product, error) {
	result, err := r.db.Exec(
		"INSERT INTO products (name, description, price) VALUES (?, ?, ?)",
		name, description, price,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &pb.Product{Id: id, Name: name, Description: description, Price: price}, nil
}
