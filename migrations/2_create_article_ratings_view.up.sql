-- Create materialized view for article ratings
CREATE MATERIALIZED VIEW materialized_articles_average_rate AS
SELECT
    article_id,
    AVG(rate) AS average_rating,
    COUNT(rate) AS total_ratings
FROM
    user_articles
WHERE
    rate > 0 
GROUP BY
    article_id;
-- WITH NO DATA; -- If we are production, we should initially created without data; then refresh at non-peak time

-- Create index for better query performance
CREATE UNIQUE INDEX idx_materialized_articles_average_rate_article_id 
ON materialized_articles_average_rate (article_id);
