package money_ports

import (
	"context"

	operation "github.com/Vovarama1992/emelya-go/internal/money/operation_model"
)

type OperationsService interface {
	ListUserOperations(ctx context.Context, userID int64) (*operation.Operations, error)
}
