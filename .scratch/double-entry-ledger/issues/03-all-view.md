# 03: Recent activity All view

**What to build:** As a user, I can open recent activity and see a descending list of transactions without mixing headers with postings or filtering in the application.

**Blocked by:** 02: Record first double-entry transaction.

**Status:** ready-for-agent

- [ ] The All view returns transaction summary items in descending business-date order.
- [ ] Each item carries description, occurred date, display amount, display account, and counterparty account.
- [ ] The query uses the transaction summary key prefix and one base-table read.
- [ ] Financial reads use strong consistency; no global secondary index is involved.
- [ ] Use-case tests with a fake repository port cover empty results and ordering.
