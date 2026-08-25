ALTER TABLE items
    ALTER COLUMN price TYPE NUMERIC(14, 2) USING price::numeric,
    ADD COLUMN price_currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    ADD COLUMN usd_exchange_rate NUMERIC(12, 4) NOT NULL DEFAULT 0,
    ADD COLUMN purchased_at DATE;

ALTER TABLE items
    ADD CONSTRAINT items_price_currency_check
        CHECK (price_currency IN ('RUB', 'USD')),
    ADD CONSTRAINT items_price_non_negative_check
        CHECK (price >= 0),
    ADD CONSTRAINT items_usd_exchange_rate_non_negative_check
        CHECK (usd_exchange_rate >= 0);
