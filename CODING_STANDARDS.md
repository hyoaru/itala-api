# Coding Standards

Judgement calls for reviewers. Mechanical checks live in `.github/workflows/verify.yml` and `lefthook.yml`.

## Slice boundaries

Each slice keeps the legacy layer shape: domain entity, application port, application use case, infrastructure adapter, and a facade for construction. Domain entities import no framework or infrastructure code. Use cases depend on ports. HTTP handlers depend on use cases.

The `account` slice owns Account and AccountNameLock. The `ledger` slice owns Transaction, Posting, Balance, TransactionSummary, and AccountStatementEntry.

No Posting repository exists. Posting has no lifecycle outside its Transaction.

## File layout

Follow the legacy feature layout exactly. One file per named thing:

- `domain/entity/` — one entity per `snake_case.go` file: `transaction.go`, `posting.go`, `account_statement_entry.go`. No consolidated model file anywhere.
- `application/port/<name>/` — `interface.go` and `errors.go`. Extra port types get their own file.
- `application/usecase/` — one use case file per action and a matching `_test.go`. Repository fakes live in `fakes_test.go`.
- `infrastructure/adapter/<name>/` — one file per concern: `dynamodb.go`, `decorated.go`, `logging.go`, `retry.go`, `idempotency.go`. Tests exist where behavior needs them: `dynamodb_test.go`, `retry_test.go`, `idempotency_test.go`. Stubs live inside the test that uses them, usually by embedding the port.
- Slice root — one facade named after the slice: `account.go`, `ledger.go`. It holds aliases and constructors only.

`domain/entity` holds domain facts only. Write infrastructure belongs in the application port or adapter. `IdempotencyRecord` lives in `application/port/ledgerrepository`, not in `domain/entity`.

Group related constructors and aliases in the facade, not in a model file. A model file defines one type.

## DynamoDB keys

Physical table keys are `PK` and `SK` in uppercase. Documentation may say `pk` and `sk` when describing logical item shape, but Go code that builds items or key conditions uses `PK` and `SK`.

Name is never part of an Account key. Account identity is UUIDv7. AccountNameLock is the only name-keyed item, and uniqueness is enforced by a conditional write in the same transaction as the account.

Build key maps inline at the call site. Prefer explicit `PK` and `SK` values over a shared key helper; the repetition is intentional. A reader should see the full key without following a helper.

## Money

Use the shared Decimal value object for every monetary amount. Never use `float64` for money.

## Error mapping

Handlers map domain errors to HTTP status at the delivery boundary:

- validation error: 400
- not found: 404
- conflict, duplicate name, concurrent modification: 409

Handlers decode JSON through the shared request decoder and write responses through the shared response writer. Authentication stays in HTTP middleware.

## Reversals

Corrections are reversals. Never edit or delete a Transaction or Posting. A reversal carries new identifiers and negated amounts on the same accounts. Read models are disposable and are written or deleted in the same transaction as the reversal.
