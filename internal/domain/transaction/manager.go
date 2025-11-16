package transaction

import "context"

// Manager интерфейс для управления транзакциями
type Manager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
