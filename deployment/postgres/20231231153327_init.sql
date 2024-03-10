-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TYPE order_status_type AS ENUM ('NEW', 'PROCESSED', 'PROCESSING', 'INVALID');

CREATE TABLE IF NOT EXISTS "user" (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    "login" varchar NOT NULL UNIQUE,
    password_hash varchar NOT NULL,
    current_balance int NOT NULL DEFAULT 0,
    withdrawn_balance int NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "order" (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    order_number varchar NOT NULL,
    status order_status_type DEFAULT 'NEW',
    accrual int NOT NULL DEFAULT 0,
    user_id uuid NOT NULL,
    created_at timestamp DEFAULT now(),
    updated_at timestamp DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT order_idx_unique_order_number UNIQUE(order_number)
);

CREATE TABLE IF NOT EXISTS operation (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    order_number varchar NOT NULL,
    amount int NOT NULL DEFAULT 0,
    user_id uuid NOT NULL,
    processed_at timestamp DEFAULT now(),
    CONSTRAINT operation_idx_unique_order_number UNIQUE(order_number)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE IF EXISTS operation;

DROP TABLE IF EXISTS "order";

DROP TABLE IF EXISTS "user";

DROP TYPE order_status_type;
-- +goose StatementEnd
