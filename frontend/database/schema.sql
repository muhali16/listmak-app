-- ====================================================
-- DATABASE SCHEMA FOR LISTMAK APPLICATION
-- ====================================================
-- Database: PostgreSQL (adjust syntax for MySQL if needed)
-- Generated: 2026-01-04
-- ====================================================

-- ====================================================
-- 1. USERS TABLE
-- (Tidak dijalankan karena sudah ada tabel login, dan login hanya menggunakan akun google)
-- Menyimpan data pengguna (OB/Admin)
-- ====================================================
-- CREATE TABLE users (
--     id SERIAL PRIMARY KEY,
--     username VARCHAR(50) UNIQUE NOT NULL,
--     email VARCHAR(100) UNIQUE NOT NULL,
--     password_hash VARCHAR(255) NOT NULL,
--     full_name VARCHAR(100),
--     role VARCHAR(20) DEFAULT 'user', -- 'admin', 'user', 'ob'
--     is_active BOOLEAN DEFAULT true,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- -- Index untuk login
-- CREATE INDEX idx_users_username ON users(username);
-- CREATE INDEX idx_users_email ON users(email);

-- ====================================================
-- 2. LISTMAK TABLE
-- Menyimpan data listmak harian
-- ====================================================
CREATE TABLE listmaks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    date DATE NOT NULL, -- Tanggal listmak (untuk daily tracking)
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    total_orders INTEGER DEFAULT 0,
    total_amount DECIMAL(12, 2) DEFAULT 0,
    paid_amount DECIMAL(12, 2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'completed', 'cancelled'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk query by date
CREATE INDEX idx_listmaks_date ON listmaks(date);
CREATE INDEX idx_listmaks_created_by ON listmaks(created_by);

-- ====================================================
-- 3. ORDERS TABLE
-- Menyimpan detail pesanan dalam listmak
-- ====================================================
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    listmak_id INTEGER NOT NULL REFERENCES listmaks(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- Nama pemesan
    order_detail TEXT NOT NULL, -- Detail pesanan
    price DECIMAL(12, 2) DEFAULT 0,
    qty INTEGER DEFAULT 1,
    total_price DECIMAL(12, 2) GENERATED ALWAYS AS (price * qty) STORED,
    is_paid BOOLEAN DEFAULT false,
    paid_at TIMESTAMP,
    added_via VARCHAR(20) DEFAULT 'parse', -- 'parse', 'manual', 'sharelink'
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk filtering
CREATE INDEX idx_orders_listmak_id ON orders(listmak_id);
CREATE INDEX idx_orders_is_paid ON orders(is_paid);
CREATE INDEX idx_orders_name ON orders(name);

-- ====================================================
-- 4. SHARE_LINKS TABLE
-- Menyimpan link share untuk input pesanan
-- ====================================================
CREATE TABLE share_links (
    id SERIAL PRIMARY KEY,
    share_id VARCHAR(20) UNIQUE NOT NULL, -- Random alphanumeric ID
    listmak_id INTEGER NOT NULL REFERENCES listmaks(id) ON DELETE CASCADE,
    title VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk lookup by share_id
CREATE UNIQUE INDEX idx_share_links_share_id ON share_links(share_id);
CREATE INDEX idx_share_links_listmak_id ON share_links(listmak_id);
CREATE INDEX idx_share_links_expires_at ON share_links(expires_at);

-- ====================================================
-- 5. VIEW_SHARES TABLE
-- Menyimpan link share untuk view-only listmak
-- ====================================================
CREATE TABLE view_shares (
    id SERIAL PRIMARY KEY,
    view_id VARCHAR(20) UNIQUE NOT NULL, -- Random alphanumeric ID
    listmak_id INTEGER NOT NULL REFERENCES listmaks(id) ON DELETE CASCADE,
    title VARCHAR(255),
    snapshot_data JSONB, -- Snapshot of listmak data at share time
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk lookup by view_id
CREATE UNIQUE INDEX idx_view_shares_view_id ON view_shares(view_id);
CREATE INDEX idx_view_shares_listmak_id ON view_shares(listmak_id);

-- ====================================================
-- 6. CONTACTS TABLE
-- Menyimpan daftar kontak
-- ====================================================
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    notes TEXT,
    is_favorite BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk query by user
CREATE INDEX idx_contacts_user_id ON contacts(user_id);
CREATE INDEX idx_contacts_name ON contacts(name);

-- ====================================================
-- 7. ORDER_HISTORY TABLE
-- Menyimpan history perubahan pesanan (audit trail)
-- ====================================================
CREATE TABLE order_history (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL, -- 'created', 'updated', 'paid', 'deleted'
    old_data JSONB,
    new_data JSONB,
    changed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_order_history_order_id ON order_history(order_id);
CREATE INDEX idx_order_history_changed_at ON order_history(changed_at);

-- ====================================================
-- 8. DAILY_SUMMARY TABLE
-- Menyimpan ringkasan harian untuk dashboard
-- ====================================================
CREATE TABLE daily_summaries (
    id SERIAL PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    total_listmaks INTEGER DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    total_amount DECIMAL(12, 2) DEFAULT 0,
    paid_amount DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_daily_summaries_date ON daily_summaries(date);

-- ====================================================
-- TRIGGER: Update listmak totals when orders change
-- ====================================================
CREATE OR REPLACE FUNCTION update_listmak_totals()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE listmaks
    SET 
        total_orders = (SELECT COUNT(*) FROM orders WHERE listmak_id = COALESCE(NEW.listmak_id, OLD.listmak_id)),
        total_amount = (SELECT COALESCE(SUM(price * qty), 0) FROM orders WHERE listmak_id = COALESCE(NEW.listmak_id, OLD.listmak_id)),
        paid_amount = (SELECT COALESCE(SUM(price * qty), 0) FROM orders WHERE listmak_id = COALESCE(NEW.listmak_id, OLD.listmak_id) AND is_paid = true),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.listmak_id, OLD.listmak_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_listmak_totals
AFTER INSERT OR UPDATE OR DELETE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_listmak_totals();

-- ====================================================
-- TRIGGER: Update timestamp on update
-- ====================================================
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_listmaks_updated_at
BEFORE UPDATE ON listmaks
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_contacts_updated_at
BEFORE UPDATE ON contacts
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- ====================================================
-- SAMPLE DATA (Optional)
-- ====================================================
-- Insert sample admin user (password: admin123)
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES ('admin', 'admin@listmak.com', '$2a$10$examplehash...', 'Administrator', 'admin');

-- Insert sample OB user (password: ob123)
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES ('ob1', 'ob1@listmak.com', '$2a$10$examplehash...', 'Office Boy 1', 'ob');
