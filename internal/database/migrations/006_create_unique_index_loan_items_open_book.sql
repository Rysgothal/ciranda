CREATE UNIQUE INDEX idx_loan_items_open_book
    ON loans_items(book_id)
    WHERE end_timestamp IS NULL;