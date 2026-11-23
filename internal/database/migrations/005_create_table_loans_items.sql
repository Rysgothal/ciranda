CREATE TABLE loans_items (
    loan_id       INTEGER NOT NULL,
    book_id       INTEGER NOT NULL,
    end_timestamp TEXT,

    PRIMARY KEY (loan_id, book_id),
    FOREIGN KEY (loan_id) REFERENCES loans(id),
    FOREIGN KEY (book_id) REFERENCES books(id)
);