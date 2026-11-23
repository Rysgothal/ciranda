CREATE TABLE sections (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (TRIM(name) <> '')
);