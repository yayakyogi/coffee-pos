CREATE TABLE IF NOT EXISTS promos (
    id VARCHAR(36) NOT NULL,
    name VARCHAR(200) NOT NULL,
    code VARCHAR(50) NOT NULL,
    type ENUM('percentage','fixed') NOT NULL,
    value BIGINT NOT NULL,
    min_order BIGINT NOT NULL DEFAULT 0,
    max_discount BIGINT NULL,
    usage_limit INT NULL,
    used_count INT NOT NULL DEFAULT 0,
    started_at DATETIME NOT NULL,
    ended_at DATETIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY idx_promos_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
