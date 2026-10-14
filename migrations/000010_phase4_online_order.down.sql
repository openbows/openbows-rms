DROP TABLE IF EXISTS online_order_extras;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_member_id;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS ck_orders_channel;
DROP INDEX IF EXISTS idx_orders_store_channel;
DROP INDEX IF EXISTS idx_orders_member_id;
DROP INDEX IF EXISTS idx_orders_channel;
ALTER TABLE orders DROP COLUMN IF EXISTS member_id;
ALTER TABLE orders DROP COLUMN IF EXISTS channel;
