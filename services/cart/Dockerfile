FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o cart-service ./services/cart/cmd

FROM gcr.io/distroless/static
COPY --from=builder /app/cart-service /cart-service
ENTRYPOINT ["/cart-service"]
