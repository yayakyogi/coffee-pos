CREATE TABLE IF NOT EXISTS shifts (
    id VARCHAR(36) NOT NULL,
    cashier_id VARCHAR(36) NOT NULL,
    opened_at DATETIME NOT NULL,
    closed_at DATETIME NULL,
    opening_cash BIGINT NOT NULL DEFAULT 0,
    closing_cash BIGINT NULL,
    total_transactions BIGINT NOT NULL DEFAULT 0,
    status ENUM('open','closed') NOT NULL DEFAULT 'open',
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_shifts_cashier_id (cashier_id),
    KEY idx_shifts_status (status),
    CONSTRAINT fk_shifts_cashier_id FOREIGN KEY (cashier_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
