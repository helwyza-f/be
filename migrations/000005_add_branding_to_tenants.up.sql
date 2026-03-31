ALTER TABLE tenants ADD COLUMN IF NOT EXISTS banner_url TEXT DEFAULT '';
-- Kita gunakan TEXT[] untuk menyimpan banyak URL foto galeri
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS gallery TEXT[] DEFAULT '{}';

ALTER TABLE resource_items ADD COLUMN is_default BOOLEAN DEFAULT false;