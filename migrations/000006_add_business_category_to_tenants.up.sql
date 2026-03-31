-- Tambah kolom business_category
ALTER TABLE tenants ADD COLUMN business_category VARCHAR(50);

-- Update data lama (opsional, jika kamu sudah punya data testing)
UPDATE tenants SET business_category = 'gaming_hub' WHERE business_category IS NULL;

-- Berikan constraint NOT NULL setelah data lama diupdate
ALTER TABLE tenants ALTER COLUMN business_category SET NOT NULL;

-- Tambahkan index agar pencarian kategori cepat
CREATE INDEX idx_tenants_business_category ON tenants(business_category);