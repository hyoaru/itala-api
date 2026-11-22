package ledgerrepository

import port "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"

func NewDecoratedLedgerRepository(inner port.LedgerRepository) port.LedgerRepository {
	return NewRetryLedgerRepository(NewLoggingLedgerRepository(inner))
}
