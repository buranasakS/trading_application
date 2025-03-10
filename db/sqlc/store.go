package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	Querier
	DeductUserBalanceTx(ctx context.Context, arg DeductUserBalanceTxParams) (DeductUserBalanceTxResult, error)
}

type PGXStore struct {
	*Queries
	db *pgx.Conn
}

func NewPGXStore(db *pgx.Conn) Store {
	return &PGXStore{
		db:      db,
		Queries: New(db),
	}
}

func NewStore(db *pgx.Conn) Store {
	return &PGXStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *PGXStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}

type DeductUserBalanceTxParams struct {
	ID      pgtype.UUID `json:"id"`
	Balance float64     `json:"balance"`
}

type DeductUserBalanceTxResult struct {
	User User `json:"user"`
}

func (store *PGXStore) DeductUserBalanceTx(ctx context.Context, arg DeductUserBalanceTxParams) (DeductUserBalanceTxResult, error) {
	var result DeductUserBalanceTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		user, err := q.GetUserDetailByID(ctx, arg.ID)
		if err != nil {
			return err
		}

		if user.Balance < arg.Balance {
			return fmt.Errorf("insufficient balance")
		}

		_, err = q.DeductUserBalance(ctx, DeductUserBalanceParams{
			ID:      arg.ID,
			Balance: arg.Balance,
		})
		if err != nil {
			return err
		}

		updatedUser, err := q.GetUserDetailByID(ctx, arg.ID)
		if err != nil {
			return err
		}

		result.User.Username = updatedUser.Username
		return nil
	})

	return result, err
}
