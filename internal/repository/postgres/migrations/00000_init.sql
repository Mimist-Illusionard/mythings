CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    image_url TEXT,

    price NUMERIC(12, 2),
    attributes JSONB NOT NULL default '{}',

    created_at TIMESTAMPZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPZ NOT NULL DEFAULT NOW(),
);

CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255), NOT NULL UNIQUE
);

CREATE TABLE items_tags (
    item_id BIGINT NOT NULL
        REFERENCES items(id)
        ON DELETE CASCADE,

    tag_id BIGINT NOT NULL
        REFERENCES tags(id)
            ON DELETE CASCADE,

    PRIMARY KEY (item_id, tag_id)
);

CREATE INDEX idx_items_tags_tag_id
    ON items_tags(tag_id);

CREATE INDEX idx_items_attributes
    ON items
    USING GIN(attributes);