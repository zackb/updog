-- Test Data for Updog Analytics
-- This file contains test data for a single domain with 3 months of pageview history

-- User
INSERT INTO users (id, email, encrypted_password, created_at, updated_at) VALUES
('Ub3teZygAo2', 'zack@bartel.com', '$2a$10$9GunO4RbMIJMDdXqfQ78ne2ZB5iOfI9nqkXzrH7Lxr7JeYutOlGvq', datetime('now', '-90 days'), datetime('now'));

-- Domain
INSERT INTO domains (id, name, user_id, verified, verification_token, created_at, updated_at) VALUES
('al41JAbrFtm', 'zackbartel.com', 'Ub3teZygAo2', 1, 'test_verification_token', datetime('now', '-90 days'), datetime('now'));

-- Countries
INSERT INTO countries (id, name) VALUES
(1, 'United States'),
(2, 'Canada'),
(3, 'United Kingdom'),
(4, 'Germany'),
(5, 'France'),
(6, 'Australia'),
(7, 'Japan'),
(8, 'Brazil'),
(9, 'India'),
(10, 'Netherlands');

-- Regions
INSERT INTO regions (id, country_id, name) VALUES
(1, 1, 'California'),
(2, 1, 'New York'),
(3, 1, 'Texas'),
(4, 1, 'Florida'),
(5, 2, 'Ontario'),
(6, 2, 'British Columbia'),
(7, 3, 'England'),
(8, 3, 'Scotland'),
(9, 4, 'Bavaria'),
(10, 4, 'Berlin'),
(11, 5, 'Île-de-France'),
(12, 6, 'New South Wales'),
(13, 7, 'Tokyo'),
(14, 8, 'São Paulo'),
(15, 9, 'Maharashtra'),
(16, 10, 'North Holland');

-- Browsers
INSERT INTO browsers (id, name) VALUES
(1, 'Chrome'),
(2, 'Firefox'),
(3, 'Safari'),
(4, 'Edge'),
(5, 'Opera'),
(6, 'Brave');

-- Operating Systems
INSERT INTO operating_systems (id, name) VALUES
(1, 'Windows'),
(2, 'macOS'),
(3, 'Linux'),
(4, 'iOS'),
(5, 'Android'),
(6, 'ChromeOS');

-- Device Types
INSERT INTO device_types (id, name) VALUES
(1, 'Desktop'),
(2, 'Mobile'),
(3, 'Tablet');

-- Languages
INSERT INTO languages (id, code) VALUES
(1, 'en'),
(2, 'en-US'),
(3, 'en-GB'),
(4, 'fr'),
(5, 'de'),
(6, 'es'),
(7, 'ja'),
(8, 'pt'),
(9, 'hi'),
(10, 'nl');

-- Referrers
INSERT INTO referrers (id, host) VALUES
(1, 'google.com'),
(2, 'twitter.com'),
(3, 'facebook.com'),
(4, 'reddit.com'),
(5, 'linkedin.com'),
(6, 'news.ycombinator.com'),
(7, 'github.com'),
(8, 'medium.com'),
(9, 'dev.to'),
(10, 'stackoverflow.com'),
(11, ''); -- direct traffic

-- Pageviews
-- Generate 3 months of data (90 days) with varying traffic patterns
-- We'll create between 50-1000 pageviews per day with 100-1000 unique visitors

-- Month 1 (Days 90-61 ago) - Lower traffic
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id) VALUES
-- Day 90 (75 pageviews, ~60 unique visitors)
('al41JAbrFtm', datetime('now', '-90 days', '+0 hours'), '/', 1, 1, 1, 1, 1, 2, 1, 1001),
('al41JAbrFtm', datetime('now', '-90 days', '+1 hours'), '/blog', 1, 1, 1, 1, 1, 2, 1, 1001),
('al41JAbrFtm', datetime('now', '-90 days', '+2 hours'), '/about', 1, 2, 2, 2, 1, 2, 11, 1002),
('al41JAbrFtm', datetime('now', '-90 days', '+3 hours'), '/', 2, 5, 3, 2, 1, 1, 1, 1003),
('al41JAbrFtm', datetime('now', '-90 days', '+4 hours'), '/projects', 3, 7, 1, 1, 1, 3, 1, 1004),
('al41JAbrFtm', datetime('now', '-90 days', '+5 hours'), '/', 1, 3, 1, 5, 2, 2, 1, 1005),
('al41JAbrFtm', datetime('now', '-90 days', '+6 hours'), '/blog/post-1', 1, 1, 1, 1, 1, 2, 1, 1006),
('al41JAbrFtm', datetime('now', '-90 days', '+7 hours'), '/', 4, 9, 2, 1, 1, 5, 2, 1007),
('al41JAbrFtm', datetime('now', '-90 days', '+8 hours'), '/contact', 1, 4, 3, 4, 2, 2, 3, 1008),
('al41JAbrFtm', datetime('now', '-90 days', '+9 hours'), '/', 5, 11, 1, 2, 1, 4, 4, 1009),
('al41JAbrFtm', datetime('now', '-90 days', '+10 hours'), '/blog', 1, 1, 1, 1, 1, 2, 1, 1010),
('al41JAbrFtm', datetime('now', '-90 days', '+11 hours'), '/', 6, 12, 3, 2, 1, 1, 1, 1011),
('al41JAbrFtm', datetime('now', '-90 days', '+12 hours'), '/projects', 1, 2, 1, 1, 1, 2, 6, 1012),
('al41JAbrFtm', datetime('now', '-90 days', '+13 hours'), '/', 7, 13, 1, 5, 2, 7, 1, 1013),
('al41JAbrFtm', datetime('now', '-90 days', '+14 hours'), '/blog/post-2', 1, 1, 2, 1, 1, 2, 7, 1014),
('al41JAbrFtm', datetime('now', '-90 days', '+15 hours'), '/', 8, 14, 1, 5, 2, 8, 8, 1015),
('al41JAbrFtm', datetime('now', '-90 days', '+16 hours'), '/about', 1, 3, 1, 1, 1, 2, 1, 1016),
('al41JAbrFtm', datetime('now', '-90 days', '+17 hours'), '/', 9, 15, 1, 5, 2, 9, 1, 1017),
('al41JAbrFtm', datetime('now', '-90 days', '+18 hours'), '/blog', 1, 1, 3, 4, 2, 2, 1, 1018),
('al41JAbrFtm', datetime('now', '-90 days', '+19 hours'), '/', 10, 16, 2, 1, 1, 10, 5, 1019),
('al41JAbrFtm', datetime('now', '-90 days', '+20 hours'), '/projects/project-1', 1, 1, 1, 1, 1, 2, 6, 1020),
('al41JAbrFtm', datetime('now', '-90 days', '+21 hours'), '/', 1, 2, 1, 2, 1, 2, 1, 1021),
('al41JAbrFtm', datetime('now', '-90 days', '+22 hours'), '/blog/post-3', 2, 6, 3, 2, 1, 1, 1, 1022),
('al41JAbrFtm', datetime('now', '-90 days', '+23 hours'), '/', 1, 1, 1, 1, 1, 2, 11, 1023),
('al41JAbrFtm', datetime('now', '-90 days', '+0 hours', '+30 minutes'), '/contact', 1, 1, 1, 1, 1, 2, 1, 1024),
('al41JAbrFtm', datetime('now', '-90 days', '+1 hours', '+15 minutes'), '/', 3, 8, 2, 1, 1, 3, 2, 1025),
('al41JAbrFtm', datetime('now', '-90 days', '+2 hours', '+45 minutes'), '/blog', 1, 4, 1, 1, 1, 2, 1, 1026),
('al41JAbrFtm', datetime('now', '-90 days', '+3 hours', '+20 minutes'), '/', 1, 1, 3, 2, 1, 1, 1, 1027),
('al41JAbrFtm', datetime('now', '-90 days', '+4 hours', '+35 minutes'), '/projects', 4, 10, 1, 1, 1, 5, 4, 1028),
('al41JAbrFtm', datetime('now', '-90 days', '+5 hours', '+50 minutes'), '/', 1, 1, 1, 5, 2, 2, 1, 1029),
('al41JAbrFtm', datetime('now', '-90 days', '+6 hours', '+10 minutes'), '/blog/post-1', 1, 1, 1, 1, 1, 2, 6, 1030),
('al41JAbrFtm', datetime('now', '-90 days', '+7 hours', '+25 minutes'), '/about', 1, 2, 2, 2, 1, 2, 1, 1031),
('al41JAbrFtm', datetime('now', '-90 days', '+8 hours', '+40 minutes'), '/', 5, 11, 1, 2, 1, 4, 1, 1032),
('al41JAbrFtm', datetime('now', '-90 days', '+9 hours', '+55 minutes'), '/projects/project-2', 1, 1, 1, 1, 1, 2, 7, 1033),
('al41JAbrFtm', datetime('now', '-90 days', '+10 hours', '+5 minutes'), '/', 1, 3, 1, 1, 1, 2, 1, 1034),
('al41JAbrFtm', datetime('now', '-90 days', '+11 hours', '+18 minutes'), '/blog', 6, 12, 3, 2, 3, 1, 1, 1035),
('al41JAbrFtm', datetime('now', '-90 days', '+12 hours', '+32 minutes'), '/', 1, 1, 1, 1, 1, 2, 8, 1036),
('al41JAbrFtm', datetime('now', '-90 days', '+13 hours', '+47 minutes'), '/contact', 7, 13, 1, 5, 2, 7, 1, 1037),
('al41JAbrFtm', datetime('now', '-90 days', '+14 hours', '+12 minutes'), '/', 1, 1, 2, 1, 1, 2, 1, 1038),
('al41JAbrFtm', datetime('now', '-90 days', '+15 hours', '+28 minutes'), '/blog/post-4', 8, 14, 1, 5, 2, 8, 9, 1039),
('al41JAbrFtm', datetime('now', '-90 days', '+16 hours', '+43 minutes'), '/', 1, 1, 1, 1, 1, 2, 1, 1040),
('al41JAbrFtm', datetime('now', '-90 days', '+17 hours', '+8 minutes'), '/projects', 9, 15, 1, 5, 2, 9, 1, 1041),
('al41JAbrFtm', datetime('now', '-90 days', '+18 hours', '+22 minutes'), '/', 1, 2, 3, 4, 2, 2, 1, 1042),
('al41JAbrFtm', datetime('now', '-90 days', '+19 hours', '+37 minutes'), '/about', 10, 16, 2, 1, 1, 10, 10, 1043),
('al41JAbrFtm', datetime('now', '-90 days', '+20 hours', '+52 minutes'), '/', 1, 1, 1, 1, 1, 2, 1, 1044),
('al41JAbrFtm', datetime('now', '-90 days', '+21 hours', '+16 minutes'), '/blog', 1, 1, 1, 2, 1, 2, 1, 1045),
('al41JAbrFtm', datetime('now', '-90 days', '+22 hours', '+31 minutes'), '/', 2, 5, 3, 2, 1, 1, 1, 1046),
('al41JAbrFtm', datetime('now', '-90 days', '+23 hours', '+46 minutes'), '/projects/project-3', 1, 1, 1, 1, 1, 2, 6, 1047),
('al41JAbrFtm', datetime('now', '-90 days', '+0 hours', '+5 minutes'), '/', 3, 7, 1, 1, 1, 3, 1, 1048),
('al41JAbrFtm', datetime('now', '-90 days', '+1 hours', '+19 minutes'), '/blog/post-5', 1, 3, 1, 5, 2, 2, 7, 1049),
('al41JAbrFtm', datetime('now', '-90 days', '+2 hours', '+34 minutes'), '/', 1, 1, 1, 1, 1, 2, 1, 1050),
('al41JAbrFtm', datetime('now', '-90 days', '+3 hours', '+49 minutes'), '/contact', 4, 9, 2, 1, 1, 5, 1, 1051),
('al41JAbrFtm', datetime('now', '-90 days', '+4 hours', '+14 minutes'), '/', 1, 4, 3, 4, 2, 2, 1, 1052),
('al41JAbrFtm', datetime('now', '-90 days', '+5 hours', '+29 minutes'), '/about', 5, 11, 1, 2, 1, 4, 1, 1053),
('al41JAbrFtm', datetime('now', '-90 days', '+6 hours', '+44 minutes'), '/', 1, 1, 1, 1, 1, 2, 1, 1054),
('al41JAbrFtm', datetime('now', '-90 days', '+7 hours', '+9 minutes'), '/blog', 6, 12, 3, 2, 1, 1, 1, 1055),
('al41JAbrFtm', datetime('now', '-90 days', '+8 hours', '+24 minutes'), '/', 1, 2, 1, 1, 1, 2, 1, 1056),
('al41JAbrFtm', datetime('now', '-90 days', '+9 hours', '+39 minutes'), '/projects', 7, 13, 1, 5, 2, 7, 1, 1057),
('al41JAbrFtm', datetime('now', '-90 days', '+10 hours', '+54 minutes'), '/', 1, 1, 2, 1, 1, 2, 1, 1058),
('al41JAbrFtm', datetime('now', '-90 days', '+11 hours', '+3 minutes'), '/blog/post-6', 8, 14, 1, 5, 2, 8, 8, 1059),
('al41JAbrFtm', datetime('now', '-90 days', '+12 hours', '+17 minutes'), '/', 1, 1, 1, 1, 1, 2, 1, 1060),
('al41JAbrFtm', datetime('now', '-89 days', '+0 hours'), '/', 1, 1, 1, 1, 1, 2, 11, 2001),
('al41JAbrFtm', datetime('now', '-89 days', '+1 hours'), '/blog', 1, 1, 1, 1, 1, 2, 1, 2002),
('al41JAbrFtm', datetime('now', '-89 days', '+2 hours'), '/about', 1, 2, 2, 2, 1, 2, 1, 2003),
('al41JAbrFtm', datetime('now', '-89 days', '+3 hours'), '/', 2, 5, 3, 2, 1, 1, 1, 2004),
('al41JAbrFtm', datetime('now', '-89 days', '+4 hours'), '/projects', 3, 7, 1, 1, 1, 3, 2, 2005),
('al41JAbrFtm', datetime('now', '-89 days', '+5 hours'), '/', 1, 3, 1, 5, 2, 2, 1, 2006),
('al41JAbrFtm', datetime('now', '-89 days', '+6 hours'), '/blog/post-1', 1, 1, 1, 1, 1, 2, 6, 2007),
('al41JAbrFtm', datetime('now', '-89 days', '+7 hours'), '/', 4, 9, 2, 1, 1, 5, 1, 2008),
('al41JAbrFtm', datetime('now', '-89 days', '+8 hours'), '/contact', 1, 4, 3, 4, 2, 2, 1, 2009),
('al41JAbrFtm', datetime('now', '-89 days', '+9 hours'), '/', 5, 11, 1, 2, 1, 4, 1, 2010),
('al41JAbrFtm', datetime('now', '-89 days', '+10 hours'), '/blog', 1, 1, 1, 1, 1, 2, 1, 2011),
('al41JAbrFtm', datetime('now', '-89 days', '+11 hours'), '/', 6, 12, 3, 2, 1, 1, 1, 2012),
('al41JAbrFtm', datetime('now', '-89 days', '+12 hours'), '/projects', 1, 2, 1, 1, 1, 2, 1, 2013),
('al41JAbrFtm', datetime('now', '-89 days', '+13 hours'), '/', 7, 13, 1, 5, 2, 7, 1, 2014),
('al41JAbrFtm', datetime('now', '-89 days', '+14 hours'), '/blog/post-2', 1, 1, 2, 1, 1, 2, 6, 2015),
('al41JAbrFtm', datetime('now', '-89 days', '+15 hours'), '/', 8, 14, 1, 5, 2, 8, 1, 2016),
('al41JAbrFtm', datetime('now', '-89 days', '+16 hours'), '/about', 1, 3, 1, 1, 1, 2, 1, 2017),
('al41JAbrFtm', datetime('now', '-89 days', '+17 hours'), '/', 9, 15, 1, 5, 2, 9, 1, 2018),
('al41JAbrFtm', datetime('now', '-89 days', '+18 hours'), '/blog', 1, 1, 3, 4, 2, 2, 1, 2019),
('al41JAbrFtm', datetime('now', '-89 days', '+19 hours'), '/', 10, 16, 2, 1, 1, 10, 1, 2020);

-- Continue with more varied data for the remaining days
-- Day 88 (150 pageviews, ~60 unique visitors - many returning visitors)
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id)
SELECT 
    'al41JAbrFtm',
    datetime('now', '-88 days', '+' || (id % 24) || ' hours', '+' || (id % 60) || ' minutes'),
    CASE (id % 8)
        WHEN 0 THEN '/'
        WHEN 1 THEN '/blog'
        WHEN 2 THEN '/about'
        WHEN 3 THEN '/projects'
        WHEN 4 THEN '/contact'
        WHEN 5 THEN '/blog/post-' || (id % 10)
        WHEN 6 THEN '/projects/project-' || (id % 5)
        ELSE '/about'
    END,
    (id % 10) + 1,
    (id % 16) + 1,
    (id % 6) + 1,
    (id % 6) + 1,
    (id % 3) + 1,
    (id % 10) + 1,
    (id % 11) + 1,
    3000 + (id % 60)  -- Reuse visitor IDs to create returning visitors
FROM (
    WITH RECURSIVE cnt(id) AS (
        SELECT 1
        UNION ALL
        SELECT id + 1 FROM cnt WHERE id < 150
    )
    SELECT id FROM cnt
);

-- Day 87 (180 pageviews, ~70 unique visitors - returning visitors)
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id)
SELECT 
    'al41JAbrFtm',
    datetime('now', '-87 days', '+' || (id % 24) || ' hours', '+' || (id % 60) || ' minutes'),
    CASE (id % 9)
        WHEN 0 THEN '/'
        WHEN 1 THEN '/blog'
        WHEN 2 THEN '/about'
        WHEN 3 THEN '/projects'
        WHEN 4 THEN '/contact'
        WHEN 5 THEN '/blog/post-' || (id % 12)
        WHEN 6 THEN '/projects/project-' || (id % 6)
        WHEN 7 THEN '/blog/category/tech'
        ELSE '/services'
    END,
    (id % 10) + 1,
    (id % 16) + 1,
    (id % 6) + 1,
    (id % 6) + 1,
    (id % 3) + 1,
    (id % 10) + 1,
    (id % 11) + 1,
    4000 + (id % 70)  -- Reuse visitor IDs
FROM (
    WITH RECURSIVE cnt(id) AS (
        SELECT 1
        UNION ALL
        SELECT id + 1 FROM cnt WHERE id < 180
    )
    SELECT id FROM cnt
);

-- Generate data for days 86-61 (Month 1 - gradually increasing traffic)
-- ~250 pageviews per day, ~100 unique visitors
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id)
SELECT 
    'al41JAbrFtm',
    datetime('now', '-' || day_offset || ' days', '+' || (id % 24) || ' hours', '+' || (id % 60) || ' minutes'),
    CASE (id % 10)
        WHEN 0 THEN '/'
        WHEN 1 THEN '/blog'
        WHEN 2 THEN '/about'
        WHEN 3 THEN '/projects'
        WHEN 4 THEN '/contact'
        WHEN 5 THEN '/blog/post-' || (id % 15)
        WHEN 6 THEN '/projects/project-' || (id % 8)
        WHEN 7 THEN '/blog/category/tech'
        WHEN 8 THEN '/services'
        ELSE '/pricing'
    END,
    (id % 10) + 1,
    (id % 16) + 1,
    (id % 6) + 1,
    (id % 6) + 1,
    (id % 3) + 1,
    (id % 10) + 1,
    (id % 11) + 1,
    (day_offset * 1000) + (id % 100)  -- Reuse visitor IDs within each day
FROM (
    WITH RECURSIVE days(day_offset) AS (
        SELECT 86
        UNION ALL
        SELECT day_offset - 1 FROM days WHERE day_offset > 61
    ),
    cnt(id) AS (
        SELECT 1
        UNION ALL
        SELECT id + 1 FROM cnt WHERE id < 250
    )
    SELECT day_offset, id FROM days, cnt
);

-- Month 2 (Days 60-31 ago) - Medium traffic
-- ~450 pageviews per day, ~200 unique visitors
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id)
SELECT 
    'al41JAbrFtm',
    datetime('now', '-' || day_offset || ' days', '+' || (id % 24) || ' hours', '+' || (id % 60) || ' minutes'),
    CASE (id % 12)
        WHEN 0 THEN '/'
        WHEN 1 THEN '/blog'
        WHEN 2 THEN '/about'
        WHEN 3 THEN '/projects'
        WHEN 4 THEN '/contact'
        WHEN 5 THEN '/blog/post-' || (id % 20)
        WHEN 6 THEN '/projects/project-' || (id % 10)
        WHEN 7 THEN '/blog/category/tech'
        WHEN 8 THEN '/blog/category/design'
        WHEN 9 THEN '/services'
        WHEN 10 THEN '/pricing'
        ELSE '/testimonials'
    END,
    (id % 10) + 1,
    (id % 16) + 1,
    (id % 6) + 1,
    (id % 6) + 1,
    (id % 3) + 1,
    (id % 10) + 1,
    (id % 11) + 1,
    (day_offset * 2000) + (id % 200)  -- Reuse visitor IDs within each day
FROM (
    WITH RECURSIVE days(day_offset) AS (
        SELECT 60
        UNION ALL
        SELECT day_offset - 1 FROM days WHERE day_offset > 31
    ),
    cnt(id) AS (
        SELECT 1
        UNION ALL
        SELECT id + 1 FROM cnt WHERE id < 450
    )
    SELECT day_offset, id FROM days, cnt
);

-- Month 3 (Days 30-1 ago) - Higher traffic
-- ~750 pageviews per day, ~300 unique visitors
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id)
SELECT 
    'al41JAbrFtm',
    datetime('now', '-' || day_offset || ' days', '+' || (id % 24) || ' hours', '+' || (id % 60) || ' minutes'),
    CASE (id % 15)
        WHEN 0 THEN '/'
        WHEN 1 THEN '/blog'
        WHEN 2 THEN '/about'
        WHEN 3 THEN '/projects'
        WHEN 4 THEN '/contact'
        WHEN 5 THEN '/blog/post-' || (id % 25)
        WHEN 6 THEN '/projects/project-' || (id % 12)
        WHEN 7 THEN '/blog/category/tech'
        WHEN 8 THEN '/blog/category/design'
        WHEN 9 THEN '/blog/category/business'
        WHEN 10 THEN '/services'
        WHEN 11 THEN '/pricing'
        WHEN 12 THEN '/testimonials'
        WHEN 13 THEN '/case-studies'
        ELSE '/faq'
    END,
    (id % 10) + 1,
    (id % 16) + 1,
    (id % 6) + 1,
    (id % 6) + 1,
    (id % 3) + 1,
    (id % 10) + 1,
    (id % 11) + 1,
    (day_offset * 5000) + (id % 300)  -- Reuse visitor IDs within each day
FROM (
    WITH RECURSIVE days(day_offset) AS (
        SELECT 30
        UNION ALL
        SELECT day_offset - 1 FROM days WHERE day_offset > 1
    ),
    cnt(id) AS (
        SELECT 1
        UNION ALL
        SELECT id + 1 FROM cnt WHERE id < 750
    )
    SELECT day_offset, id FROM days, cnt
);

-- Add some bounces (single pageview visitors) scattered throughout
-- About 20% of unique visitors should be bounces
INSERT INTO pageviews (domain_id, ts, path, country_id, region_id, browser_id, os_id, device_type_id, language_id, referrer_id, visitor_id)
SELECT 
    'al41JAbrFtm',
    datetime('now', '-' || (id % 90) || ' days', '+' || (id % 24) || ' hours', '+' || (id % 60) || ' minutes'),
    '/',  -- Bounces typically only view the landing page
    (id % 10) + 1,
    (id % 16) + 1,
    (id % 6) + 1,
    (id % 6) + 1,
    (id % 3) + 1,
    (id % 10) + 1,
    (id % 11) + 1,
    900000 + id  -- Unique visitor IDs for bounces
FROM (
    WITH RECURSIVE cnt(id) AS (
        SELECT 1
        UNION ALL
        SELECT id + 1 FROM cnt WHERE id < 500
    )
    SELECT id FROM cnt
);
