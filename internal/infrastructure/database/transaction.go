package database

import (
	trmsql "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	"github.com/exPriceD/pr-reviewer-service/internal/domain/transaction"
)

var _ transaction.Manager = (*manager.Manager)(nil)

// NewTransactionManager создает новый менеджер транзакций для PostgreSQL
func NewTransactionManager(db *PostgresDB) transaction.Manager {
	return manager.Must(
		trmsql.NewDefaultFactory(db.DB()),
	)
}
