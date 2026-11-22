# 02: Record first double-entry transaction

**What to build:** As a user, I can post a two-legged income or expense transaction and immediately read its detail. One atomic write creates the immutable write model and every read model.

**Blocked by:** 01: Create account through new account slice.

**Status:** ready-for-agent

- [ ] One `TransactWriteItems` writes the transaction header, lean postings, balance adds, one transaction summary, and one account statement entry per posting.
- [ ] Header uses `TRANSACTION#<date>#<id>`; postings use `TRANSACTION#<date>#<id>#POSTING#<id>`.
- [ ] Postings sum to zero; the application rejects a transaction that does not, and writes nothing.
- [ ] Money uses the shared Decimal value object; enum values stay lowercase.
- [ ] A written decision for the primary-account rule and the client `type`/`category` mapping is applied in this ticket.
- [ ] Detail query returns the header followed by its postings in one call.
- [ ] Each account balance equals the sum of its posting amounts after the write.
- [ ] No write-model or projection item carries global secondary index keys; financial reads use strong consistency.
- [ ] Ledger application service owns the atomic write behind an application-layer unit-of-work port.
- [ ] Use-case tests with fake ports and a fake unit of work pass. One adapter test asserts the full item set and all-or-nothing failure.
