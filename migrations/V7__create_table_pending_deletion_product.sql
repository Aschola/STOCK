-- Create pending_deletion_products table
CREATE TABLE pending_deletion_products (
    product_id INT PRIMARY KEY,
    category_name VARCHAR(255) DEFAULT NULL,
    product_name VARCHAR(255) DEFAULT NULL,
    product_code VARCHAR(50) DEFAULT NULL,
    product_description TEXT DEFAULT NULL,
    date TIMESTAMP DEFAULT NULL,
    quantity INT DEFAULT NULL,
    reorder_level INT DEFAULT NULL,
    price DECIMAL(10,2) DEFAULT NULL
);