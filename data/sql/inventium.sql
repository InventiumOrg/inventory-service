-- Sample data for local development - Milk, Condensed Milk, and Coffee products
INSERT INTO inventory (name, unit, quantity, category, location) VALUES
-- Fresh Milk Products
('Anchor Full Cream Milk 1L', 'L', 120, 'Fresh Milk', 'Cold Storage A'),
('Meadow Fresh Lite Milk 1L', 'L', 95, 'Fresh Milk', 'Cold Storage A'),
('Pura Classic Full Cream 2L', 'L', 80, 'Fresh Milk', 'Cold Storage B'),
('Dairy Farmers Low Fat Milk 1L', 'L', 110, 'Fresh Milk', 'Cold Storage A'),
('Paul''s Smarter White Milk 1L', 'L', 75, 'Fresh Milk', 'Cold Storage B'),
('A2 Platinum Full Cream 1L', 'L', 60, 'Fresh Milk', 'Cold Storage A'),

-- Condensed Milk Products
('Nestle Sweetened Condensed Milk 395g', 'g', 200, 'Condensed Milk', 'Warehouse A'),
('Carnation Sweetened Condensed Milk 397g', 'g', 180, 'Condensed Milk', 'Warehouse A'),
('Top Score Condensed Milk 390g', 'g', 150, 'Condensed Milk', 'Warehouse B'),
('Ideal Evaporated Milk 410g', 'g', 130, 'Condensed Milk', 'Warehouse A'),
('Magnolia Sweetened Condensed Milk 300g', 'g', 90, 'Condensed Milk', 'Warehouse B'),
('Eagle Brand Condensed Milk 397g', 'g', 170, 'Condensed Milk', 'Warehouse A'),
-- Condensed Milk Products
('Nestle Sweetened Condensed Milk 395g', 'box', 20, 'Condensed Milk', 'Warehouse A'),
('Carnation Sweetened Condensed Milk 397g', 'box', 180, 'Condensed Milk', 'Warehouse A'),
('Top Score Condensed Milk 390g', 'box', 50, 'Condensed Milk', 'Warehouse B'),
('Ideal Evaporated Milk 410g', 'box', 13, 'Condensed Milk', 'Warehouse A'),
('Magnolia Sweetened Condensed Milk 300g', 'box', 90, 'Condensed Milk', 'Warehouse B'),
('Eagle Brand Condensed Milk 397g', 'box', 7, 'Condensed Milk', 'Warehouse A'),

-- Coffee Products
('Nescafe Classic Instant Coffee 175g', 'g', 250, 'Coffee', 'Warehouse C'),
('Moccona Classic Medium Roast 400g', 'g', 180, 'Coffee', 'Warehouse C'),
('International Roast Coffee 500g', 'g', 200, 'Coffee', 'Warehouse D'),
('Lavazza Qualita Oro Coffee Beans 1kg', 'kg', 75, 'Coffee', 'Warehouse C'),
('Vittoria Coffee Beans Espresso 1kg', 'kg', 85, 'Coffee', 'Warehouse D'),
('Folgers Classic Roast Ground 326g', 'g', 120, 'Coffee', 'Warehouse C'),
('Maxwell House Instant Coffee 150g', 'g', 160, 'Coffee', 'Warehouse D'),
('Starbucks Pike Place Ground 340g', 'g', 95, 'Coffee', 'Warehouse C'),
('Jacobs Kronung Instant Coffee 200g', 'g', 140, 'Coffee', 'Warehouse D'),
('Douwe Egberts Pure Gold 190g', 'g', 110, 'Coffee', 'Warehouse C'),

-- UHT/Long Life Milk
('Devondale Long Life Full Cream 1L', 'L', 300, 'UHT Milk', 'Warehouse E'),
('Sanitarium So Good Soy Milk 1L', 'L', 180, 'UHT Milk', 'Warehouse E'),
('Vitasoy Almond Milk 1L', 'L', 150, 'UHT Milk', 'Warehouse E'),
('Australia''s Own Oat Milk 1L', 'L', 120, 'UHT Milk', 'Warehouse E'),

-- Premium Coffee
('Blue Mountain Coffee Beans 250g', 'g', 45, 'Premium Coffee', 'Warehouse F'),
('Kona Coffee Ground 227g', 'g', 35, 'Premium Coffee', 'Warehouse F'),
('Ethiopian Yirgacheffe Beans 500g', 'g', 60, 'Premium Coffee', 'Warehouse F');