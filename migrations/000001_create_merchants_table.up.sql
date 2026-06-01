CREATE TABLE IF NOT EXISTS merchants (
                                         id TEXT PRIMARY KEY,
                                         name TEXT NOT NULL,
                                         balance INT NOT NULL DEFAULT 0,
                                         created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );
CREATE INDEX idx_merchants_name ON merchants(name);