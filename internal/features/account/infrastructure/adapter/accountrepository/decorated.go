package accountrepository

import port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"

func NewDecoratedAccountRepository(inner port.AccountRepository) port.AccountRepository {
	return NewRetryAccountRepository(NewLoggingAccountRepository(inner))
}
