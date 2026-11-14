CREATE TABLE users (
    id       INTEGER PRIMARY KEY,
    name     TEXT NOT NULL CHECK (TRIM(name) <> ''),
    email    TEXT,
    login    TEXT UNIQUE NOT NULL CHECK (TRIM(login) <> ''),
    password TEXT NOT NULL CHECK (TRIM(password) <> '')
);

CREATE TABLE persons (
    id           INTEGER PRIMARY KEY,
    user_id      INTEGER UNIQUE NOT NULL,
    profile_name TEXT,

    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE sections (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (TRIM(name) <> '')
);

CREATE TABLE books (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL CHECK (TRIM(name) <> ''),
    author      TEXT NOT NULL CHECK (TRIM(author) <> ''),
    description TEXT,
    section_id  INTEGER NOT NULL,

    FOREIGN KEY (section_id) REFERENCES sections(id)
);

CREATE TABLE loans (
    id              INTEGER PRIMARY KEY,
    person_id       INTEGER NOT NULL,
    start_timestamp TEXT NOT NULL CHECK (TRIM(start_timestamp) <> ''),

    FOREIGN KEY (person_id) REFERENCES persons(id)
);

CREATE TABLE loans_items (
    loan_id       INTEGER NOT NULL,
    book_id       INTEGER NOT NULL,
    end_timestamp TEXT,

    PRIMARY KEY (loan_id, book_id),
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    FOREIGN KEY (book_id) REFERENCES books(id)
);

CREATE UNIQUE INDEX idx_loan_items_open_book
    ON loans_items(book_id)
    WHERE end_timestamp IS NULL;
