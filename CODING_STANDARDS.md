# Coding Standards

Judgement calls for reviewers. Mechanical checks live in `.github/workflows/verify.yml` and `lefthook.yml`.

## Slice boundaries

Each slice keeps the legacy layer shape: domain entity, application port, application use case, infrastructure adapter, and a facade for construction. Domain entities import no framework or infrastructure code. Use cases depend on ports. HTTP handlers depend on use cases.

The `account` slice owns Account and AccountNameLock. The `ledger` slice owns Transaction, Posting, Balance, TransactionSummary, and AccountStatementEntry.

No Posting repository exists. Posting has no lifecycle outside its Transaction.

## DynamoDB keys

Physical table keys are `PK` and `SK` in uppercase. Documentation may say `pk` and `sk` when describing logical item shape, but Go code that builds items or key conditions uses `PK` and `SK`.

Name is never part of an Account key. Account identity is UUIDv7. AccountNameLock is the only name-keyed item, and uniqueness is enforced by a conditional write in the same transaction as the account.

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
