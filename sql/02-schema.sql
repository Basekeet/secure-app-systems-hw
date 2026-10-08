CREATE TABLE customers (
 id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 name text NOT NULL,
 email text NOT NULL UNIQUE
);
CREATE TABLE products (
 id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 name text NOT NULL,
 price numeric(10,2) NOT NULL CHECK (price >= 0)
);
CREATE TABLE orders (
 id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 customer_id integer NOT NULL REFERENCES customers(id),
 status text NOT NULL DEFAULT 'new'
);
CREATE TABLE order_items (
 id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 order_id integer NOT NULL REFERENCES orders(id),
 product_id integer NOT NULL REFERENCES products(id),
 quantity integer NOT NULL CHECK (quantity > 0)
);
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO student;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO student;
