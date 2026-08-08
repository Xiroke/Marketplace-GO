CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT get_random_uuid(),
    token VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID references users(id) ON DELETE CASCADE NOT NULL,
    expired_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
