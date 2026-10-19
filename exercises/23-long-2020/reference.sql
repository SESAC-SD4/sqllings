SELECT title, minutes FROM titles WHERE date_added BETWEEN '2020-01-01' AND '2020-12-31' AND minutes IS NOT NULL ORDER BY minutes DESC LIMIT 3;
