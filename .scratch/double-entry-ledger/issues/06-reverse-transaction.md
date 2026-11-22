# 06: Reverse transaction

**What to build:** As a user, I can correct a mistake by reversing it. The original stays as an audit fact, my views stop showing it, and my balances return to their prior values.

**Blocked by:** 02: Record first double-entry transaction; 04: Account statement view; 05: Dashboard balances.

**Status:** ready-for-agent

- [ ] Reversal writes a new transaction and postings with negated amounts on the same accounts.
- [ ] Original header gets reversed-by identifier guarded by `attribute_not_exists`; a second reversal fails and writes nothing.
- [ ] Balances return to their pre-transaction values.
- [ ] Original summary and statement entries are deleted; the reversal gets no read model.
- [ ] Original detail still returns header and postings with a reversed marker.
- [ ] All operations commit in one `TransactWriteItems`; a forced failure writes nothing.
- [ ] Use-case tests cover the concurrency guard and projection deletion. One adapter test asserts the atomic item set.
