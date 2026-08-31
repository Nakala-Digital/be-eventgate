-- EVG-43: Add event minimum fields (title, banner, start_time, end_time, is_paid, price, quota)
ALTER TABLE events ADD COLUMN IF NOT EXISTS title VARCHAR(255);
UPDATE events SET title = name WHERE title IS NULL AND name IS NOT NULL;
ALTER TABLE events ADD COLUMN IF NOT EXISTS banner VARCHAR(255);
UPDATE events SET banner = banner_url WHERE banner IS NULL AND banner_url IS NOT NULL;
ALTER TABLE events ADD COLUMN IF NOT EXISTS start_time TIMESTAMPTZ;
UPDATE events SET start_time = start_date WHERE start_time IS NULL AND start_date IS NOT NULL;
ALTER TABLE events ADD COLUMN IF NOT EXISTS end_time TIMESTAMPTZ;
UPDATE events SET end_time = end_date WHERE end_time IS NULL AND end_date IS NOT NULL;
ALTER TABLE events ADD COLUMN IF NOT EXISTS is_paid BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE events ADD COLUMN IF NOT EXISTS price NUMERIC(12, 2) NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN IF NOT EXISTS quota INT NOT NULL DEFAULT 0;

ALTER TABLE events ALTER COLUMN name DROP NOT NULL;
ALTER TABLE events ALTER COLUMN start_date DROP NOT NULL;
ALTER TABLE events ALTER COLUMN end_date DROP NOT NULL;
