CREATE TABLE module_notification (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL,
    chat_id TEXT NOT NULL,
    message VARCHAR(1000) NOT NULL,
    condition BOOLEAN NOT NULL DEFAULT TRUE,
    cooldown INT DEFAULT 0
);