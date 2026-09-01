-- Rollback migration for category column on events table
ALTER TABLE events DROP COLUMN IF EXISTS category;
