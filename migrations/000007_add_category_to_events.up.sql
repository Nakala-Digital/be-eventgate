-- Migration to add category column to events table
ALTER TABLE events ADD COLUMN IF NOT EXISTS category VARCHAR(100) NOT NULL DEFAULT '';
