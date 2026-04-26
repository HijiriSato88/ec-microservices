package mysql

import (
	"database/sql"
	"fmt"

	"github.com/hijiri/ec-microservices/services/cart/internal/repository"
)

type mysqlRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) repository.Repository {
	return &mysqlRepo{db: db}
}

func (r *mysqlRepo) FindByUserID(userID string) (*repository.Cart, error) {
	row := r.db.QueryRow("SELECT id FROM carts WHERE user_id = ?", userID)
	var cartID int64
	if err := row.Scan(&cartID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("cart not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to scan cart: %w", err)
	}

	rows, err := r.db.Query(
		"SELECT product_id, product_name, price, quantity FROM cart_items WHERE cart_id = ?",
		cartID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	cart := &repository.Cart{UserID: userID}
	for rows.Next() {
		item := &repository.CartItem{}
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		cart.Items = append(cart.Items, item)
	}
	return cart, nil
}

func (r *mysqlRepo) Save(cart *repository.Cart) error {
	// cart が存在しなければ作成
	var cartID int64
	row := r.db.QueryRow("SELECT id FROM carts WHERE user_id = ?", cart.UserID)
	if err := row.Scan(&cartID); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to scan cart: %w", err)
		}
		result, err := r.db.Exec("INSERT INTO carts (user_id) VALUES (?)", cart.UserID)
		if err != nil {
			return fmt.Errorf("failed to insert cart: %w", err)
		}
		cartID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get cart id: %w", err)
		}
	}

	// cart_items を全削除して入れ直す
	if _, err := r.db.Exec("DELETE FROM cart_items WHERE cart_id = ?", cartID); err != nil {
		return fmt.Errorf("failed to delete cart items: %w", err)
	}
	for _, item := range cart.Items {
		_, err := r.db.Exec(
			"INSERT INTO cart_items (cart_id, product_id, product_name, price, quantity) VALUES (?, ?, ?, ?, ?)",
			cartID, item.ProductID, item.ProductName, item.Price, item.Quantity,
		)
		if err != nil {
			return fmt.Errorf("failed to insert cart item: %w", err)
		}
	}
	return nil
}
