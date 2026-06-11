ALTER TABLE users ADD COLUMN username VARCHAR(100);
UPDATE users SET username = 'user_' || id::text WHERE username IS NULL;
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);
CREATE INDEX idx_users_username ON users(username);
