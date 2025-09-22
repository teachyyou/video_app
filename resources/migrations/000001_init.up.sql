CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS videos (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    filename   text        NOT NULL,
    path       text        NOT NULL,
    size_bytes bigint      NOT NULL,
    duration_s integer     NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

