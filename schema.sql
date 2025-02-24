CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    balance DOUBLE PRECISION NOT NULL DEFAULT 0,
    affiliate_id UUID,
    FOREIGN KEY (affiliate_id) REFERENCES affiliates(id)
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price DOUBLE PRECISION NOT NULL
);

CREATE TABLE commissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    affiliate_id UUID NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    FOREIGN KEY (affiliate_id) REFERENCES affiliates(id)
);

CREATE TABLE affiliates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    master_affiliate UUID,
    balance DOUBLE PRECISION NOT NUll DEFAULT 0
);

