Status: ready-for-agent

# Double Entry Ledger Migration

## Problem Statement

As the sole user of the personal finance app, I cannot trust my balances. Today a transaction row and the account balance are written in separate DynamoDB calls. A failure between them leaves the row without its balance effect, or the balance without its row. Retrying a request can apply the same balance change twice. Edits compensate by reversing an old amount and applying a new one, but a failed compensation is discarded, so drift is silent.

I also cannot correct history cleanly. Transactions and accounts are soft-deleted, balances keep changing for deleted accounts, and no job compares a stored balance with the sum of its transactions. The four global secondary indexes sort keys differ between create and update, so ordering depends on whether a record was ever edited.

I want a model where money movement is an immutable fact, balances are rebuildable, and every read view is correct immediately after a write. I do not need multi-user features, audit history across services, phased migration, or feature flags. All production data belongs to me, so a hard cutover is acceptable.

## Solution

Replace the mutable transaction model with a double-entry ledger and synchronous projections in one DynamoDB table.

The write model holds immutable facts: Account metadata, Transaction headers, and Posting legs. The read models are disposable projections: Balance, TransactionSummary for the All view, and AccountStatementEntry for account views. Every create or reversal writes the write model and its projections in one `TransactWriteItems`, so consistency is strong and no Saga is needed.

No global secondary indexes are defined. Projection items carry purpose-built `sk` prefixes that encode each query pattern, and base-table queries can use strongly consistent reads. All items for a user live in one partition, well below the 10 GB item-collection limit at personal-finance volume.

Corrections are reversals. A reversal appends a new transaction with negated postings, adjusts balances, marks the original as reversed under a condition that allows only one reversal, and deletes the original projections. The reversal itself has no read model; the original detail view shows a reversed badge from the write model.

## User Stories

1. As a user, I want to create an account with a name and type, so that I can group my money.
2. As a user, I want account types limited to asset, liability, income, and expense, so that my reports stay consistent.
3. As a user, I want to list my accounts, so that I can see all my money buckets in one place.
4. As a user, I want to rename an account, so that I can correct a label without touching its history.
5. As a user, I want account metadata and balance stored separately, so that a name change cannot affect a financial value.
6. As a user, I want to record a two-legged transaction, so that every amount has a source and a destination.
7. As a user, I want postings to carry signed amounts, so that debits and credits are unambiguous.
8. As a user, I want postings to sum to zero, so that a transaction cannot create or destroy money.
9. As a user, I want the API to reject a transaction whose postings do not sum to zero, so that bad data never reaches storage.
10. As a user, I want the header, postings, summaries, statement entries, and balances written in one transaction, so that a failure never leaves a partial record.
11. As a user, I want a create failure to leave no items at all, so that I can retry safely.
12. As a user, I want my dashboard to show account names, types, and balances in one API call, so that the home screen loads quickly.
13. As a user, I want a recent activity list, so that I can scan what happened without opening each transaction.
14. As a user, I want the activity list to show description, date, display amount, display account, and counterparty, so that I can recognize a transaction without extra lookups.
15. As a user, I want the activity list sorted by business date, so that my timeline reads chronologically.
16. As a user, I want an account view for each account, so that I can see only that account's activity.
17. As a user, I want each statement entry to show amount, description, date, and counterparty, so that the account ledger is self-contained.
18. As a user, I want a transaction detail view, so that I can inspect the header and every posting.
19. As a user, I want detail queries to return the header before its postings, so that I can parse one result without filtering.
20. As a user, I want balances available immediately after a write, so that the dashboard never shows stale money.
21. As a user, I want base-table reads with strong consistency, so that a just-written projection is visible to the next read.
22. As a user, I want a balance rebuildable by summing postings, so that a corrupted cache can be repaired.
23. As a user, I want balance changes only through atomic add operations, so that two concurrent writes cannot overwrite each other.
24. As a user, I want no update to a posting after creation, so that history stays true.
25. As a user, I want no delete of a posting after creation, so that audit trails survive.
26. As a user, I want to reverse a transaction instead of editing it, so that the original fact remains visible.
27. As a user, I want a reversal to use new identifiers and negated amounts on the same accounts, so that the correction is itself a valid double-entry transaction.
28. As a user, I want at most one reversal per original transaction, so that I cannot accidentally reverse twice.
29. As a user, I want a reversal guarded by a condition on the original header, so that two concurrent reversal attempts cannot both succeed.
30. As a user, I want the original summary and statement entries removed on reversal, so that my activity and account views do not show a cancelled transaction.
31. As a user, I want the original detail view to remain with a reversed marker, so that I can still find the audit trail.
32. As a user, I want the reversal absent from all read views, so that I never see a correction entry as a normal transaction.
33. As a user, I want an optional fresh correction transaction, so that I can record the correct values after a reversal.
34. As a user, I want split transactions allowed, so that I can record a transaction across more than two accounts.
35. As a user, I want split transactions to show a split label instead of a counterparty, so that the UI does not lie.
36. As a user, I want categories represented by accounts, so that I do not maintain a separate category entity.
37. As a user, I want denormalized display fields only on read models, so that the write model stays pure.
38. As a user, I want read models disposable, so that they can be deleted and rebuilt from the write model.
39. As a user, I want no user profile item, so that the table contains only domain data.
40. As a user, I want a single currency assumed everywhere, so that I do not carry exchange-rate logic today.
41. As a developer, I want account and ledger slices, so that each bounded context has one owner.
42. As a developer, I want one repository for transactions and their postings, so that posting has no lifecycle outside its parent.
43. As a developer, I want a separate balance repository, so that balance projection stays in the ledger slice.
44. As a developer, I want key construction centralized, so that create and reversal cannot drift on delimiter rules.
45. As a developer, I want no global secondary indexes, so that writes are cheaper and reads are strongly consistent.
46. As a developer, I want the migration to be a hard cutover, so that no dual-write or feature-flag code remains.
47. As a developer, I want legacy account, category, and transaction code archived, so that new code lands beside a clean seam.
48. As a developer, I want shared identity and infrastructure retained, so that authentication and DynamoDB access are not rewritten.
49. As a developer, I want tests at the application service seam, so that business rules are checked without DynamoDB.
50. As a developer, I want one adapter test for the transaction item set, so that atomicity is verified without over-testing the SDK.
51. As an operator, I want item collections far under 10 GB, so that sharding is unnecessary.
52. As an operator, I want no stream, queue, or outbox in the core path, so that there are no consumers to monitor.
53. As a user, I want reversal to be auditable, so that I can see what was corrected and when.
54. As a user, I want stable lowercase enum values, so that clients and storage agree.
55. As a user, I want UUIDv7 identifiers where possible, so that identifiers sort chronologically.
56. As a developer, I want domain entities free of framework and DynamoDB code, so that business rules stay testable without AWS.
57. As a developer, I want use cases to depend on repository ports, so that storage adapters can change without touching business rules.
58. As a developer, I want HTTP handlers to depend on application use cases, so that delivery stays outside the domain.
59. As a developer, I want logging and retry decorators around adapters, so that cross-cutting behavior matches the legacy codebase.
60. As a developer, I want use cases to implement the same generic Execute contract, so that handlers stay thin and uniform.
61. As a developer, I want repository ports to own their query, page, and error types, so that the application layer controls its contracts.
62. As a developer, I want a ledger unit-of-work port, so that the atomic write stays behind an application-layer boundary.

## Implementation Decisions

- Keep one DynamoDB table partitioned by `USER#<user_id>`. Do not add a user metadata item.
- Organize the domain into two slices. The `account` slice owns Account. The `ledger` slice owns Transaction, Posting, Balance, TransactionSummary, and AccountStatementEntry.
- Use hexagonal ports and adapters inside each slice, matching the legacy codebase. Domain entities hold business rules and import no framework or infrastructure code. Application use cases depend on repository ports. DynamoDB adapters implement those ports. HTTP handlers depend on application use cases. Keep the legacy layer shape per slice: domain entity, application port, application use case, infrastructure adapter, and a slice facade for construction.
- Keep the shared generic use-case contract: `Execute(ctx, request) (response, error)`. Declare request and response types with each use case and alias them from the slice facade, as the legacy facades do. Keep repository ports in the application layer with their own query, page, and error types.
- Keep handlers thin. Decode JSON through the shared request decoder, call one use case, and write responses through the shared response writer. Keep authentication in HTTP middleware and wiring in the application composition root.
- Add one ledger application service as the transaction coordinator. Define a ledger unit-of-work port in the application layer. The DynamoDB adapter implements that port with one `TransactWriteItems`. The application service depends on the port, not on concrete adapters. No use case writes directly to DynamoDB.
- Keep the legacy repository decorator chain, outermost to innermost: retry, idempotency, logging, DynamoDB adapter. Port logging and retry now. Do not port the idempotency decorator until its storage is designed; when added, insert it in the same position. Keep the existing bounded retry policy: five attempts, 100 ms initial delay, 2 s cap.
- Reuse shared infrastructure instead of creating parallel helpers: DynamoDB client abstraction, logger, Decimal value object, and generic use-case contract. Never use float64 for money.
- Expose `AccountRepository`, `TransactionRepository`, and `BalanceRepository`. Do not create a Posting repository; Posting is not an aggregate root.
- Keep the write model pure. Account holds metadata only. Transaction holds the header, business date, timestamps, and reversal links. Posting holds identifier, transaction identifier, account identifier, and signed amount.
- Enforce account name uniqueness per user with an `AccountNameLock` item keyed by `ACCOUNT_NAME#<normalized_name>`. Normalize with trim, NFKC, then lowercase. The lock holds only `account_id` and is never read by a read path.
- Create, rename, and delete the lock in the same `TransactWriteItems` as the account. Name collision fails the whole command. Rename to the same normalized name skips lock writes. Delete frees the name.
- Keep UUID as account identity. Name is a constrained mutable label, never part of an account key.
- Do not put display fields, counterparty fields, or index keys on write-model items.
- Use these key shapes:
  - Account: `ACCOUNT#<id>`
  - Transaction: `TRANSACTION#<occurred_at>#<id>`
  - Posting: `TRANSACTION#<occurred_at>#<transaction_id>#POSTING#<id>`
  - Balance: `BALANCE#<account_id>`
  - TransactionSummary: `TRANSACTION_SUMMARY#<occurred_at>#<id>`
  - AccountStatementEntry: `ACCOUNT_STATEMENT_ENTRY#<account_id>#<occurred_at>#<id>`
- Define no global secondary indexes. Query each read model with `pk` and a `begins_with` on its `sk` prefix. Use `BatchGetItem` for dashboard account and balance pairs.
- Use `snake_case` attributes, lowercase enum values, ISO timestamps, `YYYY-MM-DD` business dates, and UUID identifiers with UUIDv7 preferred.
- Enforce the account type enum in the application layer. DynamoDB does not enforce it.
- Keep currency single and implicit.
- Treat categories as accounts. Do not create a Category entity.
- Write all six relevant item types in one `TransactWriteItems` on create: header, postings, summary, statement entries, and balance adds. Balance items are auto-created by the first add.
- Update balances only with atomic add operations. Do not set `updated_at` on a balance add.
- Use strongly consistent base-table reads for financial values.
- On reversal, append a new transaction and postings with negated amounts on the same accounts, add negated balance deltas, set `reversed_by_transaction_id` on the original under `attribute_not_exists(reversed_by_transaction_id)`, and delete the original summary and statement entries.
- Do not create read models for a reversal. Keep the reversal in the write model only.
- Enforce three-entity read rules in the application service: postings sum to zero, transaction and posting are write-once, and no write bypasses the service and repository adapter.
- Keep the legacy code archived under a clearly named legacy area until cutover, then delete it. Do not add features to the archived code.
- Switch wiring to the new slices in one cutover. Do not ship dual-write, feature flags, shadow reads, or a backfill.
- Leave the PWA request mapping unresolved. Existing clients send type and category; the new model has no category entity. Decide the account-mapping rule before handler wiring.
- Leave idempotency storage unresolved. It is deferred but required for retry safety.

## Testing Decisions

- Good tests assert external behavior only: a command produces the expected read models, balance value, reversal state, and rejection, without asserting internal helper calls.
- Primary seam: the ledger application service. Exercise it through use cases that implement the shared `Execute` contract, with fake repository ports and a fake ledger unit-of-work port. Reuse the existing use-case test style with fakes as prior art. Test create, list, detail, dashboard merge input, reversal, split handling, and validation failures here. Domain entities may have small pure unit tests, but they are not the acceptance seam.
- One adapter test: the DynamoDB adapter with a fake client. Assert a single transaction call, the full item set, the balance add operations, the conditional update on the original header, and all-or-nothing behavior on forced failure.
- Do not add HTTP-level tests. The Bruno collection is prior art for the existing contract and stays unchanged under this spec.
- Test invariants through observable outcomes:
  - A transaction with postings that do not sum to zero is rejected and writes nothing.
  - After create, balance equals the sum of posting amounts for that account.
  - After reversal, balance returns to its pre-transaction value.
  - After reversal, summary and statement entries for the original are absent.
  - Detail for the original still returns header and postings with a reversed marker.
  - A second reversal attempt fails the condition and changes nothing.
  - A create failure writes no header, posting, projection, or balance item.
  - A split transaction produces statement entries without a counterparty value.
  - No write-model item carries denormalized display fields or index keys.
- Do not test DynamoDB SDK internals, key-builder implementation details, or decorator order.
- Do not test read-model rebuild in this spec beyond asserting that balance is derivable from postings.

## Out of Scope

- Category entity and category lifecycle rules.
- Idempotency key storage, replay, and request-hash mismatch handling.
- Outbox, DynamoDB Streams, EventBridge Pipes, SQS, notifications, analytics, and search publication.
- Asynchronous projections and CQRS level two.
- Event Sourcing, event replay APIs, and time travel.
- Saga, separate services, and separate databases.
- Multi-currency, exchange rates, transfers, and recurring transactions.
- Account hierarchy, rollup reporting, payee entity, and autocomplete.
- Period close, locking, and audited financials.
- Multi-table storage, sharding, and user metadata items.
- Global secondary indexes and their definitions.
- Phased migration, backfill, dual-write, feature flags, and shadow reads.
- UI redesign and changes to the Bruno API collection.
- Audit API and full ledger history browser.

## Further Notes

- Hard cutover is accepted because there is one user and all production data belongs to that user. Preserve no legacy write path.
- No legacy data needs preserving. Purge the table before cutover; every account created by the new code has a name lock, so lock deletes require `attribute_exists(PK)`.
- The write-once rule conflicts with the reversal requirement to set `reversed_by_transaction_id` on the original header. Treat the guarded header update as the single allowed exception, or resolve the conflict before implementation.
- The primary-account rule for `display_amount` and `display_account_id` is unresolved. Need a deterministic rule before handlers are wired.
- Mapping the existing `type` and `category` request fields onto accounts is unresolved. Decide before changing the HTTP layer.
- `updated_at` on Balance is set at item creation. Atomic balance adds do not move it, per the rule that `updated_at` changes only on metadata writes.
- Branch name still says `projector-model`. The vault migration model is the active design; the branch name is stale.
- Archived legacy code may be deleted only after the new slices own all traffic and tests pass.
