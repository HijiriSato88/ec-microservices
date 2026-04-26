package repository

// Cart はカートのデータ構造
type Cart struct {
	UserID string
	Items  []*CartItem
}

// CartItem はカート内の1商品
type CartItem struct {
	ProductID   int64
	ProductName string
	Price       int64
	Quantity    int32
}

type Repository interface {
	FindByUserID(userID string) (*Cart, error)
	Save(cart *Cart) error
}
