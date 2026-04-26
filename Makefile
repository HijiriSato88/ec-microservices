.PHONY: proto migrate-product migrate-cart migrate start

proto:
	buf generate

migrate-product:
	GOOSE_DRIVER=mysql GOOSE_DBSTRING="root:password@tcp(localhost:3306)/product_db?parseTime=true" \
		goose -dir migrations/product up

migrate-cart:
	GOOSE_DRIVER=mysql GOOSE_DBSTRING="root:password@tcp(localhost:3306)/cart_db?parseTime=true" \
		goose -dir migrations/cart up

migrate:
	docker compose exec mysql mysql -uroot -ppassword -e \
		"CREATE DATABASE IF NOT EXISTS product_db; CREATE DATABASE IF NOT EXISTS cart_db;"
	$(MAKE) migrate-product
	$(MAKE) migrate-cart

start:
	docker compose up --build
