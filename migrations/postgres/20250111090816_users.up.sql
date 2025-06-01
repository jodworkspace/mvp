CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255),
    pin_hash VARCHAR(255),
    avatar_url VARCHAR(255),
    preferred_language VARCHAR(10) DEFAULT 'en',
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS links (
    issuer VARCHAR(255),
    external_id VARCHAR(255) NOT NULL,
    user_id UUID ,
    access_token VARCHAR(255),
    refresh_token VARCHAR(255),
    access_token_expires_at TIMESTAMP,
    refresh_token_expires_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    PRIMARY KEY (issuer, external_id)
)