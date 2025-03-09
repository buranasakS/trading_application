// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	AddAffiliateBalance(ctx context.Context, arg AddAffiliateBalanceParams) error
	AddUserBalance(ctx context.Context, arg AddUserBalanceParams) (int64, error)
	CheckUserExists(ctx context.Context, id pgtype.UUID) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
	CreateAffiliate(ctx context.Context, arg CreateAffiliateParams) (Affiliate, error)
	CreateCommission(ctx context.Context, arg CreateCommissionParams) (Commission, error)
	CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	DeductProductQuantity(ctx context.Context, arg DeductProductQuantityParams) (int64, error)
	DeductUserBalance(ctx context.Context, arg DeductUserBalanceParams) (int64, error)
	GetAffiliateByID(ctx context.Context, id pgtype.UUID) (Affiliate, error)
	GetAffiliateByUserID(ctx context.Context, id pgtype.UUID) (GetAffiliateByUserIDRow, error)
	GetCommissionByID(ctx context.Context, id pgtype.UUID) (Commission, error)
	GetProductByID(ctx context.Context, id pgtype.UUID) (Product, error)
	GetUserByUsernameForLogin(ctx context.Context, username string) (GetUserByUsernameForLoginRow, error)
	GetUserDetailByID(ctx context.Context, id pgtype.UUID) (GetUserDetailByIDRow, error)
	ListAffiliates(ctx context.Context) ([]Affiliate, error)
	ListCommissions(ctx context.Context) ([]Commission, error)
	ListProducts(ctx context.Context) ([]Product, error)
	ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error)
	UserBalance(ctx context.Context, id pgtype.UUID) (UserBalanceRow, error)
}

var _ Querier = (*Queries)(nil)
