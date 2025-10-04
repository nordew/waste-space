-- +goose Up
-- +goose StatementBegin
CREATE TABLE dumpster_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dumpster_id UUID NOT NULL,
    user_id UUID NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_minutes INTEGER,
    total_cost DECIMAL(10,2),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_dumpster_usages_dumpster FOREIGN KEY (dumpster_id) REFERENCES dumpsters(id) ON DELETE CASCADE,
    CONSTRAINT fk_dumpster_usages_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_dumpster_usages_status CHECK (status IN ('active', 'completed', 'cancelled')),
    CONSTRAINT chk_dumpster_usages_time CHECK (end_time IS NULL OR end_time > start_time)
);

CREATE INDEX idx_dumpster_usages_dumpster_id ON dumpster_usages(dumpster_id);
CREATE INDEX idx_dumpster_usages_user_id ON dumpster_usages(user_id);
CREATE INDEX idx_dumpster_usages_status ON dumpster_usages(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_dumpster_usages_start_time ON dumpster_usages(start_time);
CREATE INDEX idx_dumpster_usages_deleted_at ON dumpster_usages(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS dumpster_usages;
-- +goose StatementEnd
