DROP TABLE IF EXISTS article_source_categories;
DROP TABLE IF EXISTS article_source_selectors;
DROP TABLE IF EXISTS article_sources;
ALTER TABLE articles DROP COLUMN IF EXISTS source_id;
ALTER TABLE articles DROP COLUMN IF EXISTS original_url;
ALTER TABLE articles DROP COLUMN IF EXISTS is_manual;
