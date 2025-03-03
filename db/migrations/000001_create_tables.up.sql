CREATE TABLE IF NOT EXISTS affiliates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    master_affiliate UUID,
    balance DOUBLE PRECISION NOT NUll DEFAULT 0
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    password TEXT NOT  NULL,
    balance DOUBLE PRECISION NOT NULL DEFAULT 0,
    affiliate_id UUID,
    FOREIGN KEY (affiliate_id) REFERENCES affiliates(id)
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS commissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    affiliate_id UUID NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    FOREIGN KEY (affiliate_id) REFERENCES affiliates(id)
);



