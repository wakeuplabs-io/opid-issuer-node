package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/core/pagination"
	"github.com/wakeup-labs/issuer-node/internal/sqltools"
)

// NewGetAllConnectionsRequest struct
type NewGetAllConnectionsRequest struct {
	WithCredentials bool
	Query           string
	Pagination      pagination.Filter
	OrderBy         sqltools.OrderByFilters
}

// NewGetAllRequest returns the request object for obtaining all connections
func NewGetAllRequest(withCredentials *bool, query *string, page *uint, maxResults *uint, orderBy sqltools.OrderByFilters) *NewGetAllConnectionsRequest {
	var connQuery string

	if query != nil {
		connQuery = *query
	}

	pagFilter := pagination.NewFilter(maxResults, page)

	return &NewGetAllConnectionsRequest{
		WithCredentials: withCredentials != nil && *withCredentials,
		Query:           connQuery,
		Pagination:      *pagFilter,
		OrderBy:         orderBy,
	}
}

// GetStateTransactionsRequest is a request for GetStateTransitions
type GetStateTransactionsRequest struct {
	Filter     string
	Pagination pagination.Filter
	OrderBy    sqltools.OrderByFilters
}

// NewGetStateTransactionsRequest creates a new GetStateTransactionsRequest
func NewGetStateTransactionsRequest(filter string, page *uint, maxResults *uint, orderBy sqltools.OrderByFilters) *GetStateTransactionsRequest {
	pagFilter := pagination.NewFilter(maxResults, page)
	return &GetStateTransactionsRequest{
		Filter:     filter,
		Pagination: *pagFilter,
		OrderBy:    orderBy,
	}
}

// DeleteRequest struct
type DeleteRequest struct {
	ConnID            uuid.UUID
	DeleteCredentials bool
	RevokeCredentials bool
}

// NewDeleteRequest creates a new DeleteRequest
func NewDeleteRequest(connID uuid.UUID, deleteCredentials *bool, revokeCredentials *bool) *DeleteRequest {
	return &DeleteRequest{
		ConnID:            connID,
		DeleteCredentials: deleteCredentials != nil && *deleteCredentials,
		RevokeCredentials: revokeCredentials != nil && *revokeCredentials,
	}
}

// ConnectionService  is the interface implemented by the Connections service
type ConnectionService interface {
	Create(ctx context.Context, conn *domain.Connection) error
	Delete(ctx context.Context, id uuid.UUID, deleteCredentials bool, issuerDID w3c.DID) error
	DeleteCredentials(ctx context.Context, id uuid.UUID, issuerID w3c.DID) error
	GetByIDAndIssuerID(ctx context.Context, id uuid.UUID, issuerDID w3c.DID) (*domain.Connection, error)
	GetByUserID(ctx context.Context, issuerDID w3c.DID, userID w3c.DID) (*domain.Connection, error)
	GetAllByIssuerID(ctx context.Context, issuerDID w3c.DID, request *NewGetAllConnectionsRequest) ([]domain.Connection, uint, error)
	GetByUserSessionID(ctx context.Context, sessionID uuid.UUID) (*domain.Connection, error)
}
