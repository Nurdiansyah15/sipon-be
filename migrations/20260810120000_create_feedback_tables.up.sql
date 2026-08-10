-- ============================================================
-- FEEDBACK TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS feedbacks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    title            VARCHAR(200) NOT NULL,
    body             TEXT NOT NULL,
    category         VARCHAR(30) NOT NULL DEFAULT 'lainnya'
                     CHECK (category IN ('saran', 'pengaduan', 'pertanyaan', 'apresiasi', 'lainnya')),
    is_takedown      BOOLEAN NOT NULL DEFAULT false,
    takedown_reason  TEXT,
    takedown_by      UUID,
    takedown_at      TIMESTAMPTZ,
    like_count       INTEGER NOT NULL DEFAULT 0,
    comment_count    INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);
CREATE INDEX idx_feedbacks_user ON feedbacks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_feedbacks_category ON feedbacks(category) WHERE deleted_at IS NULL AND is_takedown = false;
CREATE INDEX idx_feedbacks_public ON feedbacks(created_at DESC) WHERE deleted_at IS NULL AND is_takedown = false;
CREATE INDEX idx_feedbacks_takedown ON feedbacks(is_takedown) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS feedback_attachments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feedback_id       UUID NOT NULL REFERENCES feedbacks(id) ON DELETE CASCADE,
    key               TEXT NOT NULL,
    original_filename VARCHAR(500),
    mime_type         VARCHAR(200),
    size              BIGINT,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE INDEX idx_feedback_attachments_feedback ON feedback_attachments(feedback_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS feedback_comments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feedback_id      UUID NOT NULL REFERENCES feedbacks(id) ON DELETE CASCADE,
    user_id          UUID NOT NULL,
    body             TEXT NOT NULL,
    reply_to_id      UUID,
    is_takedown      BOOLEAN NOT NULL DEFAULT false,
    takedown_reason  TEXT,
    takedown_by      UUID,
    takedown_at      TIMESTAMPTZ,
    like_count       INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    CONSTRAINT chk_comment_no_self_reply CHECK (reply_to_id IS NULL OR reply_to_id != id)
);
CREATE INDEX idx_feedback_comments_feedback ON feedback_comments(feedback_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_feedback_comments_user ON feedback_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_feedback_comments_reply ON feedback_comments(reply_to_id) WHERE deleted_at IS NULL AND reply_to_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS feedback_likes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    target_type  VARCHAR(10) NOT NULL CHECK (target_type IN ('feedback', 'comment')),
    target_id    UUID NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, target_type, target_id)
);
CREATE INDEX idx_feedback_likes_target ON feedback_likes(target_type, target_id);
CREATE INDEX idx_feedback_likes_user_target ON feedback_likes(user_id, target_type, target_id);
