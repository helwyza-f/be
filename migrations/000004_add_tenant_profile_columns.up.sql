-- Step 1: Tambahkan kolom baru ke tabel tenants
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slogan TEXT DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS address TEXT DEFAULT '';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS open_time TEXT DEFAULT '09:00';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS close_time TEXT DEFAULT '22:00';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS logo_url TEXT DEFAULT '';

-- Step 2: (Optional) Tambahkan index jika nanti kita sering cari berdasarkan slug di Landing Page
CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);