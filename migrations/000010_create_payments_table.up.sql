CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) NOT NULL,
    order_id VARCHAR(36) NOT NULL,
    method ENUM('cash','midtrans') NOT NULL,
    status ENUM('pending','paid','failed','expired') NOT NULL DEFAULT 'pending',
    amount BIGINT NOT NULL,
    midtrans_order_id VARCHAR(100) NULL,
    midtrans_token VARCHAR(500) NULL,
    midtrans_url VARCHAR(500) NULL,
    raw_notification JSON NULL,
    paid_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY idx_payments_order_id (order_id),
    UNIQUE KEY idx_payments_midtrans_order_id (midtrans_order_id),
    KEY idx_payments_status (status),
    CONSTRAINT fk_payments_order_id FOREIGN KEY (order_id) REFERENCES orders (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
