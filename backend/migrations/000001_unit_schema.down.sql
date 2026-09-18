-- Удаление индексов
DROP INDEX IF EXISTS idx_orders_restaurant_status;
DROP INDEX IF EXISTS idx_orders_created_at_desc;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_menu_items_restaurant;
DROP INDEX IF EXISTS idx_order_items_order;

-- Удаление таблиц (порядок важен: сначала дочерние, потом родительские)
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS restaurants;

-- Удаление расширения
DROP EXTENSION IF EXISTS "uuid-ossp";