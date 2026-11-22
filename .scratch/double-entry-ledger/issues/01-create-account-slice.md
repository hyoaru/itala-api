# 01: Create account through new account slice

**What to build:** As a user, I can create, read, list, and rename accounts through the API, and each operation runs through the new account slice using the same hexagonal structure as the legacy code.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [ ] Create returns identifier, name, lowercase type, and ISO timestamps.
- [ ] Account items use `ACCOUNT#<id>` under the user partition and carry no balance.
- [ ] Create writes an `ACCOUNT_NAME#<normalized_name>` lock item in the same atomic write; a duplicate name fails the whole write.
- [ ] Rename moves the lock atomically; rename to the same normalized name skips lock writes. Delete frees the name.
- [ ] List and get return only accounts for the authenticated user.
- [ ] Rename changes name and metadata timestamp only, never a financial value.
- [ ] Use case calls a repository port; DynamoDB adapter implements the port; logging and retry decorators sit in the legacy order.
- [ ] Use-case tests with fake ports pass. One adapter test asserts the account item shape.
