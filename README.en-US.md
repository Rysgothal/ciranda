# Ciranda

[Leia em Português (Brasil)](README.md)

API for managing the catalog, loans, and returns of a community library.

## Features

- Create, retrieve, update, and delete books
- Record loans and returns
- Track due dates and identify overdue loans
- List books with pagination and filters
- List overdue loans
- Validate that unavailable books cannot be borrowed

## Stack

- Go
- In development

## Expected structure (in development)

The application separates responsibilities among the handler, service, and repository.

- **Handler:** receives and responds to HTTP requests
- **Service:** applies loan and return business rules
- **Repository:** accesses data

Interfaces should be defined by the component that consumes them, without adding unnecessary layers or abstractions.

## Business rules

- A book that is already on loan cannot be borrowed again until it is returned
- Loan and return operations must preserve the book's state consistently
- Loan and return operations must run in a transaction to prevent conflicts during simultaneous access

## Quality

Business rules should be covered by service tests using fake repositories, without a direct database dependency. The system must remain consistent when simultaneous attempts are made to borrow the same book.
