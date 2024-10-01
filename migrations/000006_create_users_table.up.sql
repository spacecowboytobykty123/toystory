CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    role text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    paying_time text,
    plan text,
    balance integer NOT NULL DEFAULT 0,
    bucket text[],
    version integer NOT NULL DEFAULT 1
)