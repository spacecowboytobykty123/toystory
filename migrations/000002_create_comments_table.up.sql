CREATE TABLE IF NOT EXISTS comments (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    toy_id bigserial references toys(id) ON DELETE CASCADE,
    user_name text NOT NULL,
    text text NOT NULL,
    rating integer
);