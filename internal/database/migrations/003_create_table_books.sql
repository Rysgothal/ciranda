CREATE TABLE books (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL CHECK (TRIM(name) <> ''),
    author      TEXT NOT NULL CHECK (TRIM(author) <> ''),
    description TEXT,
    section_id  INTEGER NOT NULL,

    FOREIGN KEY (section_id) REFERENCES sections(id)
);