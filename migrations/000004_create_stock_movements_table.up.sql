CREATE TABLE IF NOT EXISTS stock_movements (
    id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    type ENUM('in','out','adjustment') NOT NULL,
    quantity INT NOT NULL,
    stock_before INT NOT NULL,
    stock_after INT NOT NULL,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_stock_movements_product_id (product_id),
    KEY idx_stock_movements_user_id (user_id),
    CONSTRAINT fk_stock_movements_product_id FOREIGN KEY (product_id) REFERENCES products (id),
    CONSTRAINT fk_stock_movements_user_id FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
