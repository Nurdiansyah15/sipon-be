DROP TRIGGER IF EXISTS trg_articles_updated_at ON articles;
DROP TRIGGER IF EXISTS trg_article_categories_updated_at ON article_categories;
DROP FUNCTION IF EXISTS trg_articles_set_updated_at();
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS article_categories;
