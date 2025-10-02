-- +goose Up
-- +goose StatementBegin
CREATE TABLE dumpsters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(50) NOT NULL,
    zip_code VARCHAR(10) NOT NULL,
    price_per_day DECIMAL(10,2) NOT NULL,
    size VARCHAR(20) NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT true,
    rating DECIMAL(3,2) DEFAULT 0.0,
    review_count INTEGER DEFAULT 0,
    capacity VARCHAR(50),
    weight VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_dumpsters_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_dumpsters_size CHECK (size IN ('small', 'medium', 'large', 'extraLarge')),
    CONSTRAINT chk_dumpsters_rating CHECK (rating >= 0 AND rating <= 5)
);

CREATE INDEX idx_dumpsters_owner_id ON dumpsters(owner_id);
CREATE INDEX idx_dumpsters_deleted_at ON dumpsters(deleted_at);
CREATE INDEX idx_dumpsters_location ON dumpsters(latitude, longitude);
CREATE INDEX idx_dumpsters_city_state ON dumpsters(city, state);
CREATE INDEX idx_dumpsters_is_available ON dumpsters(is_available) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dumpsters;
-- +goose StatementEnd
