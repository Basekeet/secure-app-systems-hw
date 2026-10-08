INSERT INTO customers (name,email) VALUES ('Anna','anna@example.com'),('Ivan','ivan@example.com');
INSERT INTO products (name,price) VALUES ('Notebook',50),('Pen',30),('Book',500);
INSERT INTO orders (customer_id,status) VALUES (1,'new'),(2,'paid');
INSERT INTO order_items (order_id,product_id,quantity) VALUES (1,1,2),(1,2,3),(2,3,1);
