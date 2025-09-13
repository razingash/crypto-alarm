CREATE TABLE module_notification_telegram_bot (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL UNIQUE,
    token TEXT NOT NULL,
    chat_id TEXT NOT NULL
);

CREATE TABLE module_notification_telegram_message (
    id BIGSERIAL PRIMARY KEY,
    element_id VARCHAR(1000) NOT NULL, -- element id from diagrams.data
    bot_id BIGINT NOT NULL REFERENCES module_notification_telegram_bot(id) ON DELETE CASCADE,
    message VARCHAR(1000) NOT NULL,
    signal BOOLEAN NOT NULL DEFAULT TRUE
);
