
-- User
INSERT INTO users (id, email, encrypted_password, name, initials, created_at, updated_at) VALUES
('Ub3teZygAo2', 'zack@bartel.com', '$2a$10$9GunO4RbMIJMDdXqfQ78ne2ZB5iOfI9nqkXzrH7Lxr7JeYutOlGvq', 'Zack Bartel', 'ZB',datetime('now'), datetime('now'));

-- Domain
INSERT INTO domains (id, name, user_id, verified, verification_token, created_at, updated_at) VALUES
('al41JAbrFtm', 'zackbartel.com', 'Ub3teZygAo2', 1, 'test_verification_token', datetime('now', '-90 days'), datetime('now'));



-- Paths
INSERT INTO paths (id, path) VALUES
(1, '/'),
(2, '/about'),
(3, '/contact'),
(4, '/projects'),
(5, '/blog'),
(6, '/blog/post-1'),
(7, '/blog/post-2'),
(8, '/blog/post-3'),
(9, '/blog/category/tech'),
(10, '/blog/category/design'),
(11, '/blog/category/business'),
(12, '/services'),
(13, '/pricing'),
(14, '/testimonials'),
(15, '/case-studies'),
(16, '/faq');

-- Browsers
INSERT INTO browsers (id, name) VALUES
(1, 'Chrome'),
(2, 'Firefox'),
(3, 'Safari'),
(4, 'Edge'),
(5, 'Opera'),
(6, 'Brave'),
(7, 'Vivaldi'),
(8, 'Samsung Internet');

-- Operating Systems
INSERT INTO operating_systems (id, name) VALUES
(1, 'Windows'),
(2, 'macOS'),
(3, 'Linux'),
(4, 'iOS'),
(5, 'Android'),
(6, 'ChromeOS'),
(7, 'Ubuntu Touch');

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
(10, 'nl'),
(11, 'zh-CN'),
(12, 'ko');

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
(10, 'Netherlands'),
(11, 'Italy'),
(12, 'Spain'),
(13, 'Mexico'),
(14, 'Russia'),
(15, 'China');

-- Regions (partial, realistic examples)
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
(12, 5, 'Provence'),
(13, 6, 'New South Wales'),
(14, 6, 'Victoria'),
(15, 7, 'Tokyo'),
(16, 7, 'Osaka'),
(17, 8, 'São Paulo'),
(18, 8, 'Rio de Janeiro'),
(19, 9, 'Maharashtra'),
(20, 9, 'Delhi'),
(21, 10, 'North Holland'),
(22, 10, 'South Holland');

-- Referrers
INSERT INTO referrers (id, host) VALUES
(1, 'google.com'),
(2, 'bing.com'),
(3, 'yahoo.com'),
(4, 'duckduckgo.com'),
(5, 'twitter.com'),
(6, 'facebook.com'),
(7, 'reddit.com'),
(8, 'linkedin.com'),
(9, 'news.ycombinator.com'),
(10, 'github.com'),
(11, 'medium.com'),
(12, 'dev.to'),
(13, 'stackoverflow.com'),
(14, 'quora.com'),
(15, 'Direct'); -- direct traffic

-- Step 1: Generate day offsets
WITH RECURSIVE days(day_offset) AS (
    SELECT 0
    UNION ALL
    SELECT day_offset + 1 FROM days WHERE day_offset < 90
),

-- Step 2: Generate traffic info per day
traffic AS (
    SELECT
        day_offset,
        date('now', '-' || day_offset || ' days') AS day,
        CASE 
            WHEN strftime('%w', date('now', '-' || day_offset || ' days')) IN ('0','6')
            THEN abs(random() % 150) + 30  -- weekend: 30-180 visitors
            ELSE abs(random() % 300) + 200 -- weekday: 200-500 visitors
        END AS unique_visitors
    FROM days
),

-- Step 3: Generate visitor sequence
visitor_seq AS (
    WITH RECURSIVE seq(i) AS (
        SELECT 1
        UNION ALL
        SELECT i + 1 FROM seq WHERE i < 1000
    )
    SELECT i FROM seq
),

-- Step 4: Assign visitors per day with session counts
visitor_sessions AS (
    SELECT
        t.day_offset,
        t.day,
        1000 + (v.i + t.day_offset * 1000) AS visitor_id,
        -- 60% bounce rate: 60% get 1 pageview, 40% get 2-5 pageviews
        CASE 
            WHEN (abs(random()) % 100) < 60 THEN 1
            ELSE (abs(random()) % 4) + 2
        END AS pageviews_in_session
    FROM traffic t
    CROSS JOIN visitor_seq v
    WHERE v.i <= t.unique_visitors
),

-- Step 5: Expand to individual pageviews
visit_num AS (
    WITH RECURSIVE nums(n) AS (
        SELECT 1
        UNION ALL
        SELECT n + 1 FROM nums WHERE n < 10
    )
    SELECT n FROM nums
),

expanded_pageviews AS (
    SELECT
        vs.day_offset,
        vs.day,
        vs.visitor_id,
        vn.n AS pageview_num,
        (abs(random()) % 15) + 1 AS path_id,
        (abs(random()) % 15) + 1 AS country_id,
        (abs(random()) % 22) + 1 AS region_id,
        (abs(random()) % 8) + 1 AS browser_id,
        (abs(random()) % 7) + 1 AS os_id,
        (abs(random()) % 4) + 1 AS device_type_id,
        (abs(random()) % 12) + 1 AS language_id,
        (abs(random()) % 15) + 1 AS referrer_id
    FROM visitor_sessions vs
    CROSS JOIN visit_num vn
    WHERE vn.n <= vs.pageviews_in_session
)

-- Step 6: Insert into pageviews
INSERT INTO pageviews (
    domain_id, ts, path_id,
    country_id, region_id, browser_id, os_id,
    device_type_id, language_id, referrer_id, visitor_id
)
SELECT
    'al41JAbrFtm',
    datetime(day || ' ' || time('00:00:00', '+' || (abs(random()) % 86400) || ' seconds')),
    path_id,
    country_id,
    region_id,
    browser_id,
    os_id,
    device_type_id,
    language_id,
    referrer_id,
    visitor_id
FROM expanded_pageviews
ORDER BY day, visitor_id, pageview_num;
