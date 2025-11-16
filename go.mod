module github.com/exPriceD/pr-reviewer-service

go 1.24.5

require github.com/lib/pq v1.10.9

require (
	github.com/avito-tech/go-transaction-manager/drivers/sql/v2 v2.0.2
	github.com/avito-tech/go-transaction-manager/trm/v2 v2.0.2
	github.com/go-chi/chi/v5 v5.2.3
	go.uber.org/mock v0.6.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
)
