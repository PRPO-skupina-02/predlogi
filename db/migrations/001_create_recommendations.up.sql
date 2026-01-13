CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    user_id UUID NOT NULL,
    movie_id UUID NOT NULL,

    reason TEXT,
    confidence_score FLOAT DEFAULT 0,

    sent_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,

    status VARCHAR(50) DEFAULT 'pending',

    generation_context JSONB,

    email_to VARCHAR(255),
    email_subject VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS idx_recommendations_user_id ON recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_status ON recommendations(status);
CREATE INDEX IF NOT EXISTS idx_recommendations_sent_at ON recommendations(sent_at);
