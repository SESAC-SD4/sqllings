ALTER TABLE titles ADD CONSTRAINT titles_minutes_check CHECK (minutes > 0);
