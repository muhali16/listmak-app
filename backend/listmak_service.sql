-- PostgreSQL schema for listmak_service
-- Generated: 2026-06-18
-- Database: listmak_service

-- Create database (run as superuser before connecting to the DB):
-- CREATE DATABASE listmak_service ENCODING 'UTF8';

--
-- Table: users
--

DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    google_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar TEXT,
    role VARCHAR(10) DEFAULT 'user',
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT idx_users_google_id UNIQUE (google_id),
    CONSTRAINT idx_users_email UNIQUE (email)
);

CREATE INDEX idx_users_deleted_at ON users (deleted_at);

--
-- Table: listmaks
--

DROP TABLE IF EXISTS listmaks CASCADE;
CREATE TABLE listmaks (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255),
    date DATE NOT NULL,
    created_by BIGINT,
    total_orders INTEGER DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0.00,
    paid_amount DECIMAL(12,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_listmaks_user FOREIGN KEY (created_by) REFERENCES users (id)
);

CREATE INDEX idx_listmaks_date ON listmaks (date);
CREATE INDEX idx_listmaks_created_by ON listmaks (created_by);
CREATE INDEX idx_listmaks_deleted_at ON listmaks (deleted_at);

--
-- Table: orders
--

DROP TABLE IF EXISTS orders CASCADE;
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    listmak_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    order_detail TEXT NOT NULL,
    price DECIMAL(12,2) DEFAULT 0.00,
    qty INTEGER DEFAULT 1,
    total_price DECIMAL(12,2) GENERATED ALWAYS AS (price * qty) STORED,
    is_paid BOOLEAN DEFAULT FALSE,
    paid_at TIMESTAMPTZ,
    added_via VARCHAR(20) DEFAULT 'parse',
    added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_orders_listmak FOREIGN KEY (listmak_id) REFERENCES listmaks (id) ON DELETE CASCADE
);

CREATE INDEX idx_orders_listmak_id ON orders (listmak_id);
CREATE INDEX idx_orders_is_paid ON orders (is_paid);
CREATE INDEX idx_orders_name ON orders (name);
CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);

--
-- Table: share_links
--

DROP TABLE IF EXISTS share_links CASCADE;
CREATE TABLE share_links (
    id BIGSERIAL PRIMARY KEY,
    share_id VARCHAR(20) NOT NULL,
    listmak_id BIGINT NOT NULL,
    title VARCHAR(255),
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_share_links_share_id UNIQUE (share_id)
);

CREATE INDEX idx_share_links_listmak_id ON share_links (listmak_id);
CREATE INDEX idx_share_links_expires_at ON share_links (expires_at);

--
-- Table: view_shares
--

DROP TABLE IF EXISTS view_shares CASCADE;
CREATE TABLE view_shares (
    id BIGSERIAL PRIMARY KEY,
    view_id VARCHAR(20) NOT NULL,
    listmak_id BIGINT NOT NULL,
    title VARCHAR(255),
    snapshot_data JSONB,
    created_by BIGINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_view_shares_view_id UNIQUE (view_id)
);

CREATE INDEX idx_view_shares_listmak_id ON view_shares (listmak_id);

--
-- Note: system_logs table is stored in a local SQLite file (logs.db), not PostgreSQL.
--

--
-- Sample data
--

INSERT INTO users (id, google_id, email, name, avatar, role, created_at, updated_at, deleted_at) VALUES
(5, '101414767172550559118', 'muhammadali55214@gmail.com', 'Muhammad Ali Mustaqim', 'https://lh3.googleusercontent.com/a/ACg8ocK_8bjXy7iyjYW9T6KyACCqVSCSkr3PvURGSvrWspxaF_3F-zPH=s96-c', 'user', '2025-12-27 17:04:07.635+07', '2025-12-27 17:04:07.635+07', NULL),
(6, '112764661582594114435', 'muhammadali55214.mri@gmail.com', 'Muhammad Ali Mustaqim MRI', 'https://lh3.googleusercontent.com/a/ACg8ocIUDlMKUur_1Hwhcxg3E_2OIu6mQVgSkL3WgzEjW3Bys1ba=s96-c', 'user', '2026-01-04 19:51:34.354+07', '2026-01-04 19:51:34.354+07', NULL);

SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

INSERT INTO listmaks (id, title, date, created_by, total_orders, total_amount, paid_amount, status, created_at, updated_at, deleted_at) VALUES
(1, 'ListMak Minggu, 4 Januari', '2026-01-04', 5, 4, 68000.00, 68000.00, 'active', '2026-01-04 22:01:49.259+07', '2026-01-04 23:43:59.953+07', NULL),
(3, 'Listmak sore', '2026-01-04', 5, 11, 0.00, 0.00, 'active', '2026-01-04 22:56:39.047+07', '2026-01-04 23:47:45.597+07', NULL),
(4, 'listjum', '2026-01-02', 5, 0, 0.00, 0.00, 'active', '2026-01-04 23:40:31.124+07', '2026-01-04 23:40:31.124+07', NULL),
(5, 'Makan siang 31', '2025-12-30', 5, 9, 85000.00, 0.00, 'active', '2026-01-04 23:52:29.350+07', '2026-01-04 23:54:20.760+07', NULL);

SELECT setval('listmaks_id_seq', (SELECT MAX(id) FROM listmaks));

INSERT INTO orders (id, listmak_id, name, order_detail, price, qty, is_paid, paid_at, added_via, added_at, updated_at, deleted_at) VALUES
(8, 1, 'Mba Ribka', 'Nasi ayam madura', 22000.00, 1, TRUE, '2026-01-04 23:43:59.665+07', 'manual', '2026-01-04 22:55:18+07', '2026-01-04 23:44:00+07', NULL),
(9, 1, 'Dimas', 'sop ayam pak min + nasi', 12000.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:01:01+07', '2026-01-04 23:07:44+07', '2026-01-04 23:15:48+07'),
(10, 1, 'Ali', 'nasi ayam kremez sambel hijau + nasi dua bungkus', 23000.00, 1, TRUE, '2026-01-04 23:43:59.867+07', 'manual', '2026-01-04 23:01:01+07', '2026-01-04 23:44:00+07', NULL),
(11, 1, 'dhani', 'cimol bojot ', 23000.00, 1, TRUE, '2026-01-04 23:43:59.786+07', 'manual', '2026-01-04 23:19:21+07', '2026-01-04 23:44:00+07', NULL),
(12, 1, 'Ichaa', 'Nasi ayam madura', 0.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:29:33+07', '2026-01-04 23:44:00+07', NULL),
(13, 3, '⁠Ali', 'nasi bu mul pake urap + tempe orek basah + 1 bakwan jagung', 0.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:30:44+07', '2026-01-04 23:30:46+07', NULL),
(14, 3, '⁠Icha', 'Sop Ayam Pak Min PAHA (Gak pake nasi)', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:30:44+07', '2026-01-04 23:30:47+07', NULL),
(15, 3, 'Safira', 'nasi ayam Madura paha', 0.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:30:44+07', '2026-01-04 23:30:46+07', NULL),
(16, 3, 'Nadiyah', 'nasi bu mul 1/2 + urap', 0.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:30:44+07', '2026-01-04 23:30:46+07', NULL),
(17, 3, 'rachel', 'sop ayam pak min PAHA + nasi', 0.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:30:44+07', '2026-01-04 23:30:46+07', NULL),
(18, 3, 'Jo', 'Sop Ayam Pak Min PAHA + nasi + tempe', 0.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:30:44+07', '2026-01-04 23:30:46+07', NULL),
(19, 3, '⁠Reni', 'nasi ayam Madura dada tidak pakai sambel', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:30:45+07', '2026-01-04 23:30:47+07', NULL),
(20, 3, 'Susan', 'nasi bu mul + pepes ikan mas', 0.00, 1, TRUE, '2026-01-04 23:47:45.574+07', 'manual', '2026-01-04 23:30:45+07', '2026-01-04 23:47:46+07', NULL),
(21, 3, 'Mona', 'nasi + kentang Mustofa (5rb) dipisah', 0.00, 1, TRUE, '2026-01-04 23:47:44.353+07', 'manual', '2026-01-04 23:30:45+07', '2026-01-04 23:47:44+07', NULL),
(22, 3, 'Icha', 'nasi goreng', 0.00, 1, TRUE, '2026-01-04 23:47:43.374+07', 'manual', '2026-01-04 23:46:58+07', '2026-01-04 23:47:43+07', NULL),
(23, 3, 'Dhani', 'ayam goreng', 0.00, 1, TRUE, '2026-01-04 23:47:41.914+07', 'manual', '2026-01-04 23:46:58+07', '2026-01-04 23:47:42+07', NULL),
(24, 5, '⁠Ali', 'nasi bu mul pake urap + tempe orek basah + 1 bakwan jagung', 20000.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:53:43+07', '2026-01-04 23:54:03+07', NULL),
(25, 5, '⁠Icha', 'Sop Ayam Pak Min PAHA (Gak pake nasi)', 40000.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:53:43+07', '2026-01-04 23:54:13+07', NULL),
(26, 5, 'Safira', 'nasi ayam Madura paha', 25000.00, 1, FALSE, NULL, 'parse', '2026-01-04 23:53:43+07', '2026-01-04 23:54:21+07', NULL),
(27, 5, 'Nadiyah', 'nasi bu mul 1/2 + urap', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:53:43+07', '2026-01-04 23:53:43+07', NULL),
(28, 5, 'rachel', 'sop ayam pak min PAHA + nasi', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:53:43+07', '2026-01-04 23:53:43+07', NULL),
(29, 5, 'Jo', 'Sop Ayam Pak Min PAHA + nasi + tempe', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:53:43+07', '2026-01-04 23:53:43+07', NULL),
(30, 5, '⁠Reni', 'nasi ayam Madura dada tidak pakai sambel', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:53:43+07', '2026-01-04 23:53:43+07', NULL),
(31, 5, 'Susan', 'nasi bu mul + pepes ikan mas', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:53:43+07', '2026-01-04 23:53:43+07', NULL),
(32, 5, 'Mona', 'nasi + kentang Mustofa (5rb) dipisah', 0.00, 1, FALSE, NULL, 'manual', '2026-01-04 23:53:43+07', '2026-01-04 23:53:43+07', NULL);

SELECT setval('orders_id_seq', (SELECT MAX(id) FROM orders));

INSERT INTO share_links (id, share_id, listmak_id, title, expires_at, is_active, created_by, created_at) VALUES
(3, 'awF8feYu', 1, 'Makan siang hari ini', '2026-01-04 22:42:00+07', TRUE, 5, '2026-01-04 22:38:17+07'),
(4, '8ztm8aV3', 1, 'Input ListMak 4/1/2026', '2026-01-04 23:20:00+07', TRUE, 5, '2026-01-04 23:19:03+07'),
(5, '0KCCq6bR', 1, 'Input ListMak 4/1/2026', '2026-01-04 12:24:00+07', TRUE, 5, '2026-01-04 23:24:37+07'),
(6, 'Z00cecbN', 1, 'Input ListMak 4/1/2026', '2026-01-04 12:27:00+07', TRUE, 5, '2026-01-04 23:27:25+07'),
(7, 'tuV6c6L6', 1, 'Input ListMak 4/1/2026', '2026-01-04 00:27:00+07', TRUE, 5, '2026-01-04 23:27:51+07'),
(8, 'Sxx9bjnR', 1, 'Input ListMak 4/1/2026', '2026-01-05 00:28:00+07', TRUE, 5, '2026-01-04 23:28:20+07'),
(9, 'NjoEtbwC', 3, 'Input ListMak 4/1/2026', '2026-01-04 23:50:00+07', TRUE, 5, '2026-01-04 23:45:11+07');

SELECT setval('share_links_id_seq', (SELECT MAX(id) FROM share_links));

SELECT setval('view_shares_id_seq', 5);
