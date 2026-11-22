# 08: Cutover and legacy deletion

**What to build:** As a developer, the new slices own all traffic. Legacy account, category, and transaction code is gone, and the service still serves the migrated flows at the existing routes.

**Blocked by:** 01, 02, 03, 04, 05, 06, 07.

**Status:** ready-for-agent

- [ ] Composition root wires only the new account and ledger slices.
- [ ] Archived legacy account, category, and transaction code is deleted.
- [ ] No legacy balance adjustment, compensation, idempotency decorator, or global secondary index key builder remains.
- [ ] Shared identity and shared infrastructure remain in place.
- [ ] Build, formatting check, vet, and all tests pass.
- [ ] Existing HTTP routes and response shapes stay stable for the migrated flows.
