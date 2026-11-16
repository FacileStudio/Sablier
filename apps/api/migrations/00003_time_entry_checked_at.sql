-- +goose Up
-- The last moment the person confirmed a running timer: a start, a resume, an edit or a
-- "keep going" answer. The idle-timer worker measures inactivity from it.

ALTER TABLE time_entries ADD COLUMN checked_at timestamp with time zone;

-- +goose Down

ALTER TABLE time_entries DROP COLUMN checked_at;
