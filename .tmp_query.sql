SELECT provider, COUNT(*) AS n FROM metadata_records GROUP BY provider ORDER BY n DESC;
SELECT '---';
SELECT 'env XUVA_TMDB_API_KEY length' AS k;
