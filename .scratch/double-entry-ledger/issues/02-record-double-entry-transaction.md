# 02: Record first double-entry transaction

**What to build:** As a user, I can post a two-legged income or expense transaction and immediately read its detail. One atomic write creates the immutable write model and every read model.

**Blocked by:** 01: Create account through new account slice.

**Status:** ready-for-agent

- [ ] One `TransactWriteItems` writes the transaction header, lean postings, balance adds, one transaction summary, and one account statement entry per posting.
- [ ] Header uses `TRANSACTION#<date>#<id>`; postings use `TRANSACTION#<date>#<id>#POSTING#<id>`.
- [ ] Postings sum to zero; the application rejects a transaction that does not, and writes nothing.
- [ ] Money uses the shared Decimal value object; enum values stay lowercase.
- [ ] `counterparty_account_id` is the category-account leg; direction comes from that account type. No `category_id` and no category lookup exist.
- [ ] Display fields use the primary-account rule: expense or income posting first, else positive side, else first posting.
- [ ] Idempotency-Key is required. Result item is written in the same atomic write; replay returns the stored transaction ID and a different payload returns 409.
- [ ] Detail query returns the header followed by its postings in one call.
- [ ] Each account balance equals the sum of its posting amounts after the write.
- [ ] No write-model or projection item carries global secondary index keys; financial reads use strong consistency.
- [ ] Ledger application service owns the atomic write behind an application-layer unit-of-work port.
- [ ] Use-case tests with fake ports and a fake unit of work pass. One adapter test asserts the full item set and all-or-nothing failure.
