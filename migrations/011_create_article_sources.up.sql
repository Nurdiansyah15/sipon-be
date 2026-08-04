ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS source_id   UUID,
    ADD COLUMN IF NOT EXISTS original_url VARCHAR(1000),
    ADD COLUMN IF NOT EXISTS is_manual   BOOLEAN NOT NULL DEFAULT TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_original_url
    ON articles (original_url)
    WHERE original_url IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_articles_source_id
    ON articles (source_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS article_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(50)  NOT NULL UNIQUE,
    name            VARCHAR(100) NOT NULL,
    base_url        VARCHAR(500) NOT NULL,
    auto_publish    BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    last_scraped_at TIMESTAMPTZ,
    created_by      UUID,
    updated_by      UUID,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_article_sources_active
    ON article_sources (is_active)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS article_source_selectors (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id        UUID NOT NULL UNIQUE,
    content_selector TEXT,
    author_selector  TEXT,
    tags_selector    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_source_selectors_source FOREIGN KEY (source_id)
        REFERENCES article_sources (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS article_source_categories (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id           UUID NOT NULL,
    category_key        VARCHAR(100) NOT NULL,
    url_suffix          VARCHAR(500),
    url_override        VARCHAR(500),
    article_limit       INTEGER NOT NULL DEFAULT 10,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    article_category_id UUID,
    keywords            JSONB,
    last_scraped_at     TIMESTAMPTZ,
    updated_by          UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_source_categories_source FOREIGN KEY (source_id)
        REFERENCES article_sources (id) ON DELETE CASCADE,
    CONSTRAINT fk_source_categories_article_category FOREIGN KEY (article_category_id)
        REFERENCES article_categories (id) ON DELETE SET NULL,
    CONSTRAINT uq_source_category UNIQUE (source_id, category_key)
);

CREATE INDEX IF NOT EXISTS idx_source_categories_source_active
    ON article_source_categories (source_id, is_active)
    WHERE is_active = TRUE;
