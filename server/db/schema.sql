CREATE TABLE IF NOT EXISTS spotify_auth (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    scope         TEXT NOT NULL,
    expiry        INTEGER NOT NULL,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO spotify_auth (id, access_token, refresh_token, scope, expiry)
SELECT 1, '', '', '', 0
WHERE NOT EXISTS (SELECT 1 FROM spotify_auth WHERE id = 1);
