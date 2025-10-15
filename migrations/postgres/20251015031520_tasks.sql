-- +goose Up
CREATE TABLE IF NOT EXISTS tasks (
                                     id UUID PRIMARY KEY,
                                     title VARCHAR(255) NOT NULL,
                                     details TEXT,
                                     is_completed BOOLEAN DEFAULT FALSE,
                                     priority INT,
                                     start_date TIMESTAMP,
                                     due_date TIMESTAMP,
                                     owner_id UUID NOT NULL,
                                     created_at TIMESTAMP,
                                     updated_at TIMESTAMP,
                                     FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS tasks;

-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
