ALTER TABLE titles ADD CONSTRAINT titles_country_check CHECK (country NOT LIKE ',%');
