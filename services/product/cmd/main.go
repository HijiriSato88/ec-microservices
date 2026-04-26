package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	_ "github.com/go-sql-driver/mysql"
	pb "github.com/hijiri/ec-microservices/gen/go/product"
	grpcgateway "github.com/hijiri/ec-microservices/services/product/internal/gateway/grpc"
	"github.com/hijiri/ec-microservices/services/product/internal/repository/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/product_db?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	r := mysql.NewRepository(db)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterProductServiceServer(srv, grpcgateway.NewProductServer(r))
	reflection.Register(srv)

	log.Printf("product listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
