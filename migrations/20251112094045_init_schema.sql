-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'user')),
    image_path VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL CHECK (type IN ('direct', 'group')),
    name VARCHAR(255),
    creator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chats_type ON chats(type);
CREATE INDEX idx_chats_creator_id ON chats(creator_id);
CREATE INDEX idx_chats_created_at ON chats(created_at DESC);

CREATE TABLE chat_participants (
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_read_message_id INTEGER,
    last_read_at TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_participants_user_id ON chat_participants(user_id);
CREATE INDEX idx_chat_participants_joined_at ON chat_participants(joined_at DESC);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP
);

CREATE INDEX idx_messages_chat_id ON messages(chat_id, sent_at DESC);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_sent_at ON messages(sent_at DESC);

-- Foreign keys
ALTER TABLE chat_participants
    ADD CONSTRAINT fk_chat_participants_last_read_message
    FOREIGN KEY (last_read_message_id)
    REFERENCES messages(id)
    ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS chat_participants CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
