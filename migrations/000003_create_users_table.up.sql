CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password TEXT NOT NULL, -- Hashed password
    role VARCHAR(20) DEFAULT 'owner', -- 'owner', 'staff'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(email) -- Email harus unik di seluruh platform
);

-- Index untuk mempercepat login
CREATE INDEX idx_users_email ON users(email);