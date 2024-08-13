-- Migration script to create the products table

CREATE TABLE products (
    product_id INT AUTO_INCREMENT PRIMARY KEY,    
    category_name VARCHAR(100),                       
    product_name VARCHAR(100),                         
    product_code VARCHAR(100),                        
    product_description VARCHAR(255),                  
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,          
    quantity INT,                                      
    reorder_level INT,                                
    price DOUBLE(10,2)                                
);


CREATE INDEX idx_product_code ON products (product_code);
