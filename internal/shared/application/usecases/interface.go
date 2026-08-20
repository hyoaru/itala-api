package usecases

import "context"

type UseCase[T any] interface {
	Execute(ctx context.Context) T
}
