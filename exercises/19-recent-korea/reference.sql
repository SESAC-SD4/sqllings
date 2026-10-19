SELECT title, date_added FROM titles WHERE country = 'South Korea' AND date_added IS NOT NULL ORDER BY date_added DESC LIMIT 3;
