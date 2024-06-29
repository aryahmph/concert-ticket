CREATE TYPE ORDER_STATUS AS ENUM ('created', 'completed', 'cancelled');

CREATE TABLE IF NOT EXISTS orders
(
    id          VARCHAR(26) PRIMARY KEY,
    category_id SMALLINT     NOT NULL,
    email       VARCHAR(255) NOT NULL,
    va_code     VARCHAR(255),
    status      ORDER_STATUS DEFAULT 'created',
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_orders_unique ON orders (email, status) WHERE NOT (status = 'cancelled');

CREATE TABLE IF NOT EXISTS tickets
(
    id          SERIAL PRIMARY KEY,
    order_id    VARCHAR(26),
    category_id SMALLINT NOT NULL,
    row         SMALLINT NOT NULL,
    "column"    SMALLINT NOT NULL,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tickets_order_id ON tickets (order_id);
CREATE INDEX idx_tickets_category_order ON tickets (category_id, order_id);