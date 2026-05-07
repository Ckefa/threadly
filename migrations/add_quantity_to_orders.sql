-- Add quantity column to orders table
ALTER TABLE orders ADD COLUMN quantity INTEGER DEFAULT 1;

-- Update existing orders to have quantity 1 if no order items exist
UPDATE orders SET quantity = 1 WHERE quantity IS NULL;
