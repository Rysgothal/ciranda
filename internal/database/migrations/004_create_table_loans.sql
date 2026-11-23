CREATE TABLE loans (
    id              INTEGER PRIMARY KEY,
    users_id        INTEGER NOT NULL,
    start_timestamp TEXT NOT NULL CHECK (TRIM(start_timestamp) <> ''),

    FOREIGN KEY (users_id) REFERENCES users(id)
);