CREATE TABLE users (
    id       INTEGER PRIMARY KEY,
    name     TEXT NOT NULL CHECK (TRIM(name) <> ''),
    email    TEXT,
    login    TEXT UNIQUE NOT NULL CHECK (TRIM(login) <> ''),
    password TEXT NOT NULL CHECK (TRIM(password) <> '')
);