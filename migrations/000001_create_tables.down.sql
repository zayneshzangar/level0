-- Dropping indexes
DROP INDEX IF EXISTS idx_items_order_uid;
DROP INDEX IF EXISTS idx_payment_order_uid;
DROP INDEX IF EXISTS idx_delivery_order_uid;

-- Dropping tables
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS orders;
