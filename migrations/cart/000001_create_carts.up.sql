CREATE TABLE IF NOT EXISTS carts (
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id    VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cart_items (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    cart_id      BIGINT NOT NULL,
    product_id   BIGINT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price        BIGINT NOT NULL,
    quantity     INT NOT NULL,
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    UNIQUE KEY uq_cart_product (cart_id, product_id)
);
