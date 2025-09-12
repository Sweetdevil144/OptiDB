-- OptiDB Demo Data with Intentional Performance Problems
-- Creates realistic tables with performance bottlenecks for testing

-- Users table (intentionally missing index on email)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active'
);

-- Orders table (intentionally missing index on user_id and status)
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    order_date TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'pending',
    total_amount DECIMAL(10,2),
    shipping_address TEXT
);

-- Order items (intentionally missing composite index)
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL,
    product_name VARCHAR(255),
    quantity INTEGER,
    price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Events table (for correlated subquery problems)
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    event_type VARCHAR(50),
    event_data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert demo users
INSERT INTO users (email, first_name, last_name, status) VALUES
('john.doe@email.com', 'John', 'Doe', 'active'),
('jane.smith@email.com', 'Jane', 'Smith', 'active'),
('bob.johnson@email.com', 'Bob', 'Johnson', 'active'),
('alice.brown@email.com', 'Alice', 'Brown', 'inactive'),
('charlie.wilson@email.com', 'Charlie', 'Wilson', 'active'),
('diana.davis@email.com', 'Diana', 'Davis', 'active'),
('frank.miller@email.com', 'Frank', 'Miller', 'pending'),
('grace.taylor@email.com', 'Grace', 'Taylor', 'active'),
('henry.anderson@email.com', 'Henry', 'Anderson', 'active'),
('ivy.thomas@email.com', 'Ivy', 'Thomas', 'inactive'),
('jack.jackson@email.com', 'Jack', 'Jackson', 'active'),
('kelly.white@email.com', 'Kelly', 'White', 'active'),
('liam.harris@email.com', 'Liam', 'Harris', 'active'),
('mia.martin@email.com', 'Mia', 'Martin', 'pending'),
('noah.garcia@email.com', 'Noah', 'Garcia', 'active'),
('olivia.martinez@email.com', 'Olivia', 'Martinez', 'active'),
('peter.robinson@email.com', 'Peter', 'Robinson', 'active'),
('quinn.clark@email.com', 'Quinn', 'Clark', 'inactive'),
('ruby.rodriguez@email.com', 'Ruby', 'Rodriguez', 'active'),
('sam.lewis@email.com', 'Sam', 'Lewis', 'active'),
('tina.lee@email.com', 'Tina', 'Lee', 'active'),
('uma.walker@email.com', 'Uma', 'Walker', 'pending'),
('victor.hall@email.com', 'Victor', 'Hall', 'active'),
('wendy.allen@email.com', 'Wendy', 'Allen', 'active'),
('xavier.young@email.com', 'Xavier', 'Young', 'active'),
('yara.hernandez@email.com', 'Yara', 'Hernandez', 'inactive'),
('zoe.king@email.com', 'Zoe', 'King', 'active'),
('adam.wright@email.com', 'Adam', 'Wright', 'active'),
('bella.lopez@email.com', 'Bella', 'Lopez', 'active'),
('carlos.hill@email.com', 'Carlos', 'Hill', 'active');

-- Insert demo orders (spread across users)
INSERT INTO orders (user_id, status, total_amount, shipping_address) VALUES
(1, 'completed', 149.99, '123 Main St, City, State 12345'),
(2, 'pending', 89.50, '456 Oak Ave, Town, State 67890'),
(3, 'shipped', 299.00, '789 Pine Rd, Village, State 11111'),
(1, 'completed', 75.25, '123 Main St, City, State 12345'),
(4, 'cancelled', 199.99, '321 Elm St, Borough, State 22222'),
(5, 'completed', 45.00, '654 Maple Dr, County, State 33333'),
(2, 'processing', 120.75, '456 Oak Ave, Town, State 67890'),
(6, 'completed', 350.00, '987 Cedar Ln, District, State 44444'),
(7, 'pending', 25.99, '147 Birch St, Area, State 55555'),
(8, 'shipped', 189.99, '258 Spruce Ave, Region, State 66666'),
(3, 'completed', 99.99, '789 Pine Rd, Village, State 11111'),
(9, 'processing', 275.50, '369 Willow Way, Zone, State 77777'),
(10, 'completed', 159.00, '741 Aspen Ct, Sector, State 88888'),
(1, 'pending', 89.99, '123 Main St, City, State 12345'),
(11, 'shipped', 199.50, '852 Poplar Pl, Division, State 99999'),
(12, 'completed', 125.00, '963 Hickory Rd, Territory, State 10101'),
(5, 'cancelled', 75.75, '654 Maple Dr, County, State 33333'),
(13, 'processing', 245.99, '159 Walnut St, Locality, State 20202'),
(14, 'completed', 89.00, '357 Chestnut Ave, Municipality, State 30303'),
(15, 'pending', 169.99, '468 Sycamore Dr, Township, State 40404'),
(2, 'completed', 55.50, '456 Oak Ave, Town, State 67890'),
(16, 'shipped', 299.99, '579 Magnolia Ln, Parish, State 50505'),
(17, 'processing', 134.25, '681 Dogwood Ct, Precinct, State 60606'),
(18, 'completed', 89.99, '792 Redwood Way, Ward, State 70707'),
(19, 'cancelled', 199.00, '814 Sequoia Pl, District, State 80808'),
(20, 'completed', 145.50, '925 Cypress Rd, Community, State 90909'),
(3, 'pending', 75.99, '789 Pine Rd, Village, State 11111'),
(21, 'shipped', 220.00, '136 Fir St, Neighborhood, State 12121'),
(22, 'completed', 99.25, '247 Palm Ave, Vicinity, State 13131'),
(23, 'processing', 179.99, '358 Bamboo Dr, Locality, State 14141');

-- Insert order items
INSERT INTO order_items (order_id, product_name, quantity, price) VALUES
(1, 'Laptop Stand', 1, 49.99),
(1, 'Wireless Mouse', 2, 25.00),
(1, 'USB Cable', 3, 15.00),
(2, 'Coffee Mug', 1, 12.50),
(2, 'Notebook Set', 2, 38.50),
(3, 'Monitor', 1, 299.00),
(4, 'Keyboard', 1, 75.25),
(5, 'Webcam', 1, 199.99),
(6, 'Headphones', 1, 45.00),
(7, 'Phone Case', 1, 15.75),
(7, 'Screen Protector', 2, 12.50),
(7, 'Charging Cable', 3, 8.00),
(8, 'Tablet', 1, 350.00),
(9, 'Power Bank', 1, 25.99),
(10, 'Bluetooth Speaker', 1, 89.99),
(10, 'Carrying Case', 1, 35.00),
(10, 'Memory Card', 2, 32.50),
(11, 'Smart Watch', 1, 99.99),
(12, 'Fitness Tracker', 1, 75.50),
(12, 'Wall Charger', 2, 12.00),
(12, 'Car Mount', 1, 25.00),
(13, 'Gaming Mouse', 1, 59.00),
(13, 'Mouse Pad', 1, 15.00),
(13, 'Cable Organizer', 3, 8.50),
(14, 'Desk Lamp', 1, 45.00),
(14, 'Pen Holder', 1, 12.99),
(14, 'Sticky Notes', 5, 3.00),
(15, 'External Drive', 1, 89.99),
(15, 'USB Hub', 1, 35.00),
(15, 'Laptop Sleeve', 1, 25.00),
(16, 'Wireless Earbuds', 1, 149.99),
(17, 'Phone Stand', 1, 19.99),
(17, 'Wireless Charger', 1, 29.99),
(17, 'Cable Set', 1, 24.99),
(18, 'Portable Monitor', 1, 199.50),
(19, 'Gaming Headset', 1, 89.99),
(20, 'Mechanical Keyboard', 1, 134.25),
(21, 'Webcam Light', 1, 35.00),
(21, 'Microphone', 1, 55.50),
(22, 'Tablet Stand', 1, 29.99),
(22, 'Stylus Pen', 2, 15.00),
(23, 'Laptop Cooler', 1, 45.50),
(23, 'Privacy Screen', 1, 35.25),
(24, 'Document Scanner', 1, 199.00),
(25, 'Label Printer', 1, 145.50),
(26, 'Surge Protector', 1, 35.99),
(26, 'Extension Cord', 1, 19.99),
(26, 'Cable Clips', 10, 1.50),
(27, 'Monitor Arm', 1, 89.99),
(27, 'Desk Organizer', 1, 25.00),
(28, 'Wireless Keyboard', 1, 65.00);

-- Insert events for each user
INSERT INTO events (user_id, event_type, event_data) VALUES
(1, 'login', '{"ip": "192.168.1.1", "device": "desktop"}'),
(1, 'page_view', '{"page": "/products", "duration": 45}'),
(1, 'purchase', '{"order_id": 1, "amount": 149.99}'),
(2, 'login', '{"ip": "192.168.1.2", "device": "mobile"}'),
(2, 'search', '{"query": "coffee mug", "results": 25}'),
(2, 'purchase', '{"order_id": 2, "amount": 89.50}'),
(3, 'login', '{"ip": "192.168.1.3", "device": "tablet"}'),
(3, 'page_view', '{"page": "/monitors", "duration": 120}'),
(3, 'purchase', '{"order_id": 3, "amount": 299.00}'),
(4, 'login', '{"ip": "192.168.1.4", "device": "desktop"}'),
(4, 'cart_add', '{"product": "webcam", "price": 199.99}'),
(4, 'cart_abandon', '{"reason": "price_too_high"}'),
(5, 'login', '{"ip": "192.168.1.5", "device": "mobile"}'),
(5, 'purchase', '{"order_id": 6, "amount": 45.00}'),
(6, 'login', '{"ip": "192.168.1.6", "device": "desktop"}'),
(6, 'page_view', '{"page": "/tablets", "duration": 90}'),
(7, 'signup', '{"referrer": "google", "campaign": "summer_sale"}'),
(7, 'login', '{"ip": "192.168.1.7", "device": "mobile"}'),
(8, 'login', '{"ip": "192.168.1.8", "device": "desktop"}'),
(8, 'search', '{"query": "bluetooth speaker", "results": 15}'),
(9, 'login', '{"ip": "192.168.1.9", "device": "tablet"}'),
(9, 'wishlist_add', '{"product": "smart_watch", "price": 99.99}'),
(10, 'login', '{"ip": "192.168.1.10", "device": "mobile"}'),
(10, 'purchase', '{"order_id": 13, "amount": 159.00}'),
(11, 'login', '{"ip": "192.168.1.11", "device": "desktop"}'),
(12, 'signup', '{"referrer": "facebook", "campaign": "back_to_school"}'),
(13, 'login', '{"ip": "192.168.1.13", "device": "mobile"}'),
(14, 'login', '{"ip": "192.168.1.14", "device": "desktop"}'),
(15, 'login', '{"ip": "192.168.1.15", "device": "tablet"}'),
(16, 'signup', '{"referrer": "instagram", "campaign": "influencer"}'),
(17, 'login', '{"ip": "192.168.1.17", "device": "mobile"}'),
(18, 'login', '{"ip": "192.168.1.18", "device": "desktop"}'),
(19, 'login', '{"ip": "192.168.1.19", "device": "tablet"}'),
(20, 'login', '{"ip": "192.168.1.20", "device": "mobile"}');

-- Create some problematic queries to generate slow query patterns

-- Query 1: Seq scan on users (missing index on email)
SELECT * FROM users WHERE email = 'john.doe@email.com';

-- Query 2: Seq scan on orders (missing index on user_id)
SELECT * FROM orders WHERE user_id = 1;

-- Query 3: Bad join without proper indexes
SELECT u.first_name, u.last_name, o.total_amount 
FROM users u 
JOIN orders o ON u.id = o.user_id 
WHERE o.status = 'completed';

-- Query 4: Correlated subquery (inefficient)
SELECT u.*, 
       (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) as order_count
FROM users u;

-- Query 5: Missing composite index
SELECT * FROM orders 
WHERE user_id = 1 AND status = 'completed' 
ORDER BY order_date DESC;

-- Query 6: Inefficient aggregation
SELECT u.first_name, u.last_name, SUM(oi.price * oi.quantity) as total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN order_items oi ON o.id = oi.order_id
GROUP BY u.id, u.first_name, u.last_name;

-- Query 7: Text search without index
SELECT * FROM order_items WHERE product_name LIKE '%Mouse%';

-- Query 8: JSON query without GIN index
SELECT * FROM events WHERE event_data->>'device' = 'mobile';

-- Query 9: Complex join with multiple tables
SELECT u.email, o.order_date, oi.product_name, e.event_type
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN order_items oi ON o.id = oi.order_id
LEFT JOIN events e ON u.id = e.user_id
WHERE u.status = 'active' AND o.status = 'completed';

-- Query 10: Inefficient count with join
SELECT COUNT(DISTINCT o.id) as order_count
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE u.status = 'active';

-- Run these queries multiple times to generate statistics
SELECT pg_stat_statements_reset();

-- Execute slow queries multiple times to build up stats
DO $$
DECLARE
    i INTEGER;
BEGIN
    FOR i IN 1..10 LOOP
        PERFORM * FROM users WHERE email = 'john.doe@email.com';
        PERFORM * FROM orders WHERE user_id = 1;
        PERFORM u.first_name, u.last_name, o.total_amount 
        FROM users u JOIN orders o ON u.id = o.user_id 
        WHERE o.status = 'completed';
        
        PERFORM u.*, (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) as order_count
        FROM users u;
        
        PERFORM * FROM orders WHERE user_id = 1 AND status = 'completed' ORDER BY order_date DESC;
        PERFORM * FROM order_items WHERE product_name LIKE '%Mouse%';
        PERFORM * FROM events WHERE event_data->>'device' = 'mobile';
    END LOOP;
END $$;
