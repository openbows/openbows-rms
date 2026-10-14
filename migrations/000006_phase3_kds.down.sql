DROP TABLE IF EXISTS kds_ticket_items;
DROP TABLE IF EXISTS kds_tickets;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS ck_order_items_kitchen_status;
DROP INDEX IF EXISTS idx_order_items_kitchen_status;
ALTER TABLE order_items DROP COLUMN IF EXISTS kitchen_status;
