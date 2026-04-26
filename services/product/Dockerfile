FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o product-service ./services/product/cmd

FROM gcr.io/distroless/static
COPY --from=builder /app/product-service /product-service
ENTRYPOINT ["/product-service"]
