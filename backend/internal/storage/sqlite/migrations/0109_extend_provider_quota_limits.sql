-- +goose Up
ALTER TABLE quota_limits ADD COLUMN used_value REAL;
ALTER TABLE quota_limits ADD COLUMN limit_state TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE quota_limits DROP COLUMN limit_state;
ALTER TABLE quota_limits DROP COLUMN used_value;
