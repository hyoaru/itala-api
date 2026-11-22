# 05: Dashboard balances

**What to build:** As a user, I can open the dashboard and see account names, types, and up-to-date balances in one call.

**Blocked by:** 02: Record first double-entry transaction.

**Status:** ready-for-agent

- [ ] The dashboard returns account metadata merged with its balance in one API response.
- [ ] The application queries account items and batch-gets balance items, then merges in the application layer.
- [ ] Balances reflect postings immediately after a write.
- [ ] Balance items are created by the first atomic add; balance adds do not move the metadata timestamp.
- [ ] Balance reads use strong consistency.
- [ ] Use-case tests with fake ports cover merge, empty accounts, and a missing balance item.
