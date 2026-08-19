-- +goose Up
CREATE TYPE product_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,

    status product_status NOT NULL DEFAULT 'draft',

    creator_id UUID NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id),

    deleted_at TIMESTAMPTZ DEFAULT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category_id ON products(category_id);

CREATE INDEX idx_products_catalog
ON products (created_at DESC)
WHERE deleted_at IS NULL AND status = 'published';

CREATE TRIGGER set_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_timestamp();

-- +goose Down
DROP TRIGGER set_products_updated_at ON products;
DROP TABLE products;
DROP TYPE product_status;
