CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) DEFAULT NULL,
    password VARCHAR(255) DEFAULT NULL,
    full_name VARCHAR(255) DEFAULT NULL,
    org_name VARCHAR(255) DEFAULT NULL,
    role_name VARCHAR(255) DEFAULT NULL,
    organization_id BIGINT DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    created_by BIGINT DEFAULT NULL,
    phonenumber BIGINT DEFAULT NULL,
    INDEX idx_users_deleted_at (deleted_at),
    INDEX idx_users_organization_id (organization_id)
);