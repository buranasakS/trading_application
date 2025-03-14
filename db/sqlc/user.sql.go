// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addUserBalance = `-- name: AddUserBalance :execrows
UPDATE users SET balance = balance + $1 WHERE id = $2
`

type AddUserBalanceParams struct {
	Balance float64     `json:"balance"`
	ID      pgtype.UUID `json:"id"`
}

func (q *Queries) AddUserBalance(ctx context.Context, arg AddUserBalanceParams) (int64, error) {
	result, err := q.db.Exec(ctx, addUserBalance, arg.Balance, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const checkUserExists = `-- name: CheckUserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
`

func (q *Queries) CheckUserExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	row := q.db.QueryRow(ctx, checkUserExists, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const countUsers = `-- name: CountUsers :one
SELECT COUNT(*) FROM users
`

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (username, password, affiliate_id) VALUES ($1, $2, $3) RETURNING id, username, password, balance, affiliate_id
`

type CreateUserParams struct {
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	AffiliateID pgtype.UUID `json:"affiliate_id"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Username, arg.Password, arg.AffiliateID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Balance,
		&i.AffiliateID,
	)
	return i, err
}

const deductUserBalance = `-- name: DeductUserBalance :execrows
UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1
`

type DeductUserBalanceParams struct {
	Balance float64     `json:"balance"`
	ID      pgtype.UUID `json:"id"`
}

func (q *Queries) DeductUserBalance(ctx context.Context, arg DeductUserBalanceParams) (int64, error) {
	result, err := q.db.Exec(ctx, deductUserBalance, arg.Balance, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getUserByUsernameForLogin = `-- name: GetUserByUsernameForLogin :one
SELECT id, username, password, affiliate_id, balance
FROM users
WHERE username = $1
LIMIT 1
`

type GetUserByUsernameForLoginRow struct {
	ID          pgtype.UUID `json:"id"`
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	AffiliateID pgtype.UUID `json:"affiliate_id"`
	Balance     float64     `json:"balance"`
}

func (q *Queries) GetUserByUsernameForLogin(ctx context.Context, username string) (GetUserByUsernameForLoginRow, error) {
	row := q.db.QueryRow(ctx, getUserByUsernameForLogin, username)
	var i GetUserByUsernameForLoginRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.AffiliateID,
		&i.Balance,
	)
	return i, err
}

const getUserDetailByID = `-- name: GetUserDetailByID :one
SELECT id, username, balance, affiliate_id FROM users WHERE id = $1
`

type GetUserDetailByIDRow struct {
	ID          pgtype.UUID `json:"id"`
	Username    string      `json:"username"`
	Balance     float64     `json:"balance"`
	AffiliateID pgtype.UUID `json:"affiliate_id"`
}

func (q *Queries) GetUserDetailByID(ctx context.Context, id pgtype.UUID) (GetUserDetailByIDRow, error) {
	row := q.db.QueryRow(ctx, getUserDetailByID, id)
	var i GetUserDetailByIDRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Balance,
		&i.AffiliateID,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, username, balance, affiliate_id
FROM users
ORDER BY id
LIMIT $1 OFFSET $2
`

type ListUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListUsersRow struct {
	ID          pgtype.UUID `json:"id"`
	Username    string      `json:"username"`
	Balance     float64     `json:"balance"`
	AffiliateID pgtype.UUID `json:"affiliate_id"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]ListUsersRow, error) {
	rows, err := q.db.Query(ctx, listUsers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUsersRow{}
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Balance,
			&i.AffiliateID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const userBalance = `-- name: UserBalance :one
SELECT id, balance FROM users WHERE id = $1
`

type UserBalanceRow struct {
	ID      pgtype.UUID `json:"id"`
	Balance float64     `json:"balance"`
}

func (q *Queries) UserBalance(ctx context.Context, id pgtype.UUID) (UserBalanceRow, error) {
	row := q.db.QueryRow(ctx, userBalance, id)
	var i UserBalanceRow
	err := row.Scan(&i.ID, &i.Balance)
	return i, err
}
