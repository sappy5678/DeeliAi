-- Table: users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for users
CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_email ON users (email);

-- Table: articles
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    url TEXT UNIQUE NOT NULL,
    title TEXT,
    description TEXT,
    image_url TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for articles
CREATE INDEX idx_articles_url ON articles (url);

-- Table: user_articles
CREATE TABLE user_articles (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    score SMALLINT CHECK (score >= 1 AND score <= 5),
    collected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (user_id, article_id)
);

-- Index for user_articles
-- CREATE INDEX idx_user_articles_user_id ON user_articles (user_id);
CREATE INDEX idx_user_articles_article_id ON user_articles (article_id);
CREATE INDEX idx_user_articles_user_article_deleted_at ON user_articles (user_id, article_id, deleted_at);


-- Table: metadata_fetch_retries
CREATE TABLE metadata_fetch_retries (
    id BIGSERIAL PRIMARY KEY,
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    retry_count SMALLINT DEFAULT 0 NOT NULL,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    next_attempt_at TIMESTAMP WITH TIME ZONE,
    status SMALLINT NOT NULL DEFAULT 0, -- 0: pending, 1: success, 2: failed
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for metadata_fetch_retries
CREATE INDEX idx_metadata_fetch_retries_article_id ON metadata_fetch_retries (article_id);
CREATE INDEX idx_metadata_fetch_retries_next_attempt_status ON metadata_fetch_retries (next_attempt_at, status);