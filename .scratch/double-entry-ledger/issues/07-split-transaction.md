# 07: Split transaction

**What to build:** As a user, I can record one transaction across more than two accounts. Each affected account shows its own leg, and the activity view marks the transaction as a split.

**Blocked by:** 02: Record first double-entry transaction; 04: Account statement view.

**Status:** ready-for-agent

- [ ] A transaction with more than two posting legs is accepted when postings sum to zero.
- [ ] Summary and statement entries carry no counterparty account for a split.
- [ ] The API response exposes a split marker so the UI can show a split label.
- [ ] Each account balance reflects only its own posting amount.
- [ ] All items and balance adds still commit in one atomic write.
- [ ] Use-case tests cover a three-leg split and rejection of an unbalanced split.
