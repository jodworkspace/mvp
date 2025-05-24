CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);


CREATE TABLE IF NOT EXISTS federated_users (
    user_id UUID ,
    issuer VARCHAR(255),
    external_id VARCHAR(255) NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id),
    PRIMARY KEY (user_id, issuer)
)