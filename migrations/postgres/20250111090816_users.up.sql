CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);


CREATE TABLE IF NOT EXISTS federated_users
(
    user_id UUID PRIMARY KEY,
    external_id VARCHAR(255),
    FOREIGN KEY (user_id) REFERENCES users(id)
)