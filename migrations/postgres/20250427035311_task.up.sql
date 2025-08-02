CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    details TEXT,
    is_completed BOOLEAN DEFAULT FALSE,
    priority_level INT,
    start_date TIMESTAMP,
    estimated_duration BIGINT, -- store duration in nanoseconds
    due_date TIMESTAMP,
    owner_id UUID NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id)
);