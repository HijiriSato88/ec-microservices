package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	_ "github.com/go-sql-driver/mysql"
	pb "github.com/hijiri/ec-microservices/gen/go/cart"
	grpcgateway "github.com/hijiri/ec-microservices/services/cart/internal/gateway/grpc"
	"github.com/hijiri/ec-microservices/services/cart/internal/repository/mysql"
	"google.golang.org/grpc"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/cart_db?parseTime=true"
	}

	productAddr := os.Getenv("PRODUCT_ADDR")
	if productAddr == "" {
		productAddr = "localhost:50051"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	productClient, err := grpcgateway.NewProductClient(productAddr)
	if err != nil {
		log.Fatalf("failed to connect to product-service: %v", err)
	}

	r := mysql.NewRepository(db)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterCartServiceServer(srv, grpcgateway.NewCartServer(r, productClient))
	// reflection.Register(srv)

	log.Printf("cart listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
