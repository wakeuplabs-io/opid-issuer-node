package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgx/v4"

	"github.com/wakeup-labs/issuer-node/internal/common"
	"github.com/wakeup-labs/issuer-node/internal/core/domain"
	"github.com/wakeup-labs/issuer-node/internal/core/ports"
	"github.com/wakeup-labs/issuer-node/internal/db"
	"github.com/wakeup-labs/issuer-node/internal/repositories"
)

// ErrConnectionDoesNotExist connection does not exist
var ErrConnectionDoesNotExist = errors.New("connection does not exist")

type connection struct {
	connRepo   ports.ConnectionRepository
	claimsRepo ports.ClaimRepository
	storage    *db.Storage
}

// NewConnection returns a new connection service
func NewConnection(connRepo ports.ConnectionRepository, claimsRepo ports.ClaimRepository, storage *db.Storage) ports.ConnectionService {
	return &connection{
		connRepo:   connRepo,
		claimsRepo: claimsRepo,
		storage:    storage,
	}
}

func (c *connection) Create(ctx context.Context, connection *domain.Connection) error {
	return c.storage.Pgx.BeginFunc(ctx, func(tx pgx.Tx) error {
		_, err := c.connRepo.Save(ctx, c.storage.Pgx, connection)
		return err
	})
}

func (c *connection) Delete(ctx context.Context, id uuid.UUID, deleteCredentials bool, issuerDID w3c.DID) error {
	return c.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			if deleteCredentials {
				err := c.deleteCredentials(ctx, id, issuerDID, tx)
				if err != nil {
					return err
				}
			}
			return c.delete(ctx, id, issuerDID, tx)
		})
}

func (c *connection) DeleteCredentials(ctx context.Context, id uuid.UUID, issuerID w3c.DID) error {
	return c.deleteCredentials(ctx, id, issuerID, c.storage.Pgx)
}

func (c *connection) GetByIDAndIssuerID(ctx context.Context, id uuid.UUID, issuerDID w3c.DID) (*domain.Connection, error) {
	conn, err := c.connRepo.GetByIDAndIssuerID(ctx, c.storage.Pgx, id, issuerDID)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return conn, nil
}

func (c *connection) GetByUserSessionID(ctx context.Context, sessionID uuid.UUID) (*domain.Connection, error) {
	conn, err := c.connRepo.GetByUserSessionID(ctx, c.storage.Pgx, sessionID)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return conn, nil
}

func (c *connection) GetByUserID(ctx context.Context, issuerDID w3c.DID, userID w3c.DID) (*domain.Connection, error) {
	conn, err := c.connRepo.GetByUserID(ctx, c.storage.Pgx, issuerDID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return conn, nil
}

func (c *connection) GetAllByIssuerID(ctx context.Context, issuerDID w3c.DID, filter *ports.NewGetAllConnectionsRequest) ([]domain.Connection, uint, error) {
	conns, count, err := c.connRepo.GetAllWithCredentialsByIssuerID(ctx, c.storage.Pgx, issuerDID, filter)
	if filter.WithCredentials {
		for i := range conns {
			claims, err := c.claimsRepo.GetClaimsOfAConnection(ctx, c.storage.Pgx, issuerDID, conns[i].UserDID)
			if err != nil {
				return nil, 0, err
			}
			conns[i].Credentials = common.ToPointer(domain.Credentials(claims))
		}

		return conns, count, err
	}
	return conns, count, err
}

func (c *connection) delete(ctx context.Context, id uuid.UUID, issuerDID w3c.DID, pgx db.Querier) error {
	err := c.connRepo.Delete(ctx, pgx, id, issuerDID)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return ErrConnectionDoesNotExist
		}
		return err
	}

	return nil
}

func (c *connection) deleteCredentials(ctx context.Context, id uuid.UUID, issuerID w3c.DID, pgx db.Querier) error {
	return c.connRepo.DeleteCredentials(ctx, pgx, id, issuerID)
}
