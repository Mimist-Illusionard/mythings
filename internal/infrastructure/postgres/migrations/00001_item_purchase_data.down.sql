ALTER TABLE items
    DROP CONSTRAINT IF EXISTS items_usd_exchange_rate_non_negative_check,
    DROP CONSTRAINT IF EXISTS items_price_non_negative_check,
    DROP CONSTRAINT IF EXISTS items_price_currency_check,
    DROP COLUMN IF EXISTS purchased_at,
    DROP COLUMN IF EXISTS usd_exchange_rate,
    DROP COLUMN IF EXISTS price_currency,
    ALTER COLUMN price TYPE BIGINT USING ROUND(price)::bigint;
