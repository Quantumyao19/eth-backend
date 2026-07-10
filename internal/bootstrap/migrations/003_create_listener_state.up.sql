CREATE TABLE listener_state (
    name varchar(50) PRIMARY KEY,
    last_processed_block BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);