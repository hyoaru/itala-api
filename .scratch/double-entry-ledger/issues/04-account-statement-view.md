# 04: Account statement view

**What to build:** As a user, I can open an account and see only that account's activity, with amount, description, date, and counterparty already present.

**Blocked by:** 02: Record first double-entry transaction.

**Status:** ready-for-agent

- [ ] The account view returns statement entries in descending business-date order.
- [ ] Each entry carries description, transaction occurred date, signed amount, and counterparty account.
- [ ] The query uses the account statement key prefix scoped to one account.
- [ ] Each posting produces exactly one statement entry; accounts do not leak into each other's views.
- [ ] Use-case tests with a fake repository port cover empty results and account isolation.
