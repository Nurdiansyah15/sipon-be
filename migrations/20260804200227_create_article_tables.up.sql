CREATE TABLE IF NOT EXISTS article_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(120) NOT NULL UNIQUE,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_article_categories_slug
    ON article_categories (slug)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS articles (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title         VARCHAR(300) NOT NULL,
    content       TEXT NOT NULL,
    summary       TEXT,
    category_id   UUID REFERENCES article_categories (id) ON DELETE SET NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'draft'
                  CHECK (status IN ('draft', 'published', 'archived')),
    author        VARCHAR(200) NOT NULL,
    thumbnail_url TEXT,
    view_count    INTEGER NOT NULL DEFAULT 0,
    is_featured   BOOLEAN NOT NULL DEFAULT FALSE,
    created_by    UUID,
    updated_by    UUID,
    published_at  TIMESTAMPTZ,
    archived_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_articles_status_published
    ON articles (status, published_at DESC)
    WHERE status = 'published' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_articles_category
    ON articles (category_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_articles_featured
    ON articles (is_featured, published_at DESC)
    WHERE is_featured = TRUE AND status = 'published' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_articles_created_by
    ON articles (created_by)
    WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION trg_articles_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_articles_updated_at ON articles;
CREATE TRIGGER trg_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION trg_articles_set_updated_at();

DROP TRIGGER IF EXISTS trg_article_categories_updated_at ON article_categories;
CREATE TRIGGER trg_article_categories_updated_at
    BEFORE UPDATE ON article_categories
    FOR EACH ROW
    EXECUTE FUNCTION trg_articles_set_updated_at();
