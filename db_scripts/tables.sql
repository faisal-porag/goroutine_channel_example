CREATE TABLE vouchers (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL,
    min_order_amount double precision NOT NULL,
    discount_amount double precision,
    discount_percentage INT,
    max_discount_amount double precision
);