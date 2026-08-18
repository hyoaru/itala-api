# Access Patterns

Personal finance app. Every transaction carries a category (income vs. expense, e.g. "groceries").

## Read patterns

| Pattern                                                     | Note                                                                      |
| ----------------------------------------------------------- | ------------------------------------------------------------------------- |
| List all accounts of the user with balances                 | Read pattern: list, ordered by `name`. Attributes: id, name, balance      |
| Read a single account                                       | Read pattern: point lookup                                                |
| Read transaction history for a single account paginated     | Read pattern: newest first, paginated                                     |
| Filter transactions by transaction type                     | Read pattern: newest first, paginated                                     |
| Filter transactions by category                             | Read pattern: newest first, paginated                                     |
| Filter transactions by date range                           | Read pattern: newest first, paginated                                     |
| List all transactions across all accounts for a single user | Read pattern: newest first, paginated                                     |
| Monthly report of totals by category                        | Computed after fetch: sum amount grouped by category (income vs. expense) |
| Search by description                                       | Read pattern: full scan                                                   |

## Write patterns

| Pattern                   | Note                                                                                                                                                        |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Create a transaction      | Write pattern: insert. Income or expense, with category, amount, date, description. Keep account balance in sync                                            |
| Update a transaction      | Write pattern: point update (amount, category, description, date). If date/account changes it is a re-write (delete + insert). Keep account balance in sync |
| Delete a transaction      | Write pattern: point delete. Keep account balance in sync                                                                                                   |
| Transfer between accounts | Write pattern: atomic debit + credit (two balance changes and one transaction)                                                                              |

