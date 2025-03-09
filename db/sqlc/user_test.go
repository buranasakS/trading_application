package db

import (
	"context"
	"testing"

	"github.com/buranasakS/trading_application/helpers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	affiliate := createRandomAffiliate(t)
	hashPassword, err := helpers.HashedPassword(uuid.New().String())
	require.NoError(t, err)
	arg := CreateUserParams{
		Username:    uuid.New().String(),
		Password:    hashPassword,
		AffiliateID: affiliate.ID,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.AffiliateID, user.AffiliateID)

	require.NotZero(t, user.ID)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestAddUserBalance(t *testing.T) {
	user := createRandomUser(t)

	arg := AddUserBalanceParams{
		ID:      user.ID,
		Balance: 100,
	}

	affectedRows, err := testQueries.AddUserBalance(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), affectedRows)

	userBalance, err := testQueries.UserBalance(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, arg.Balance, userBalance.Balance)
}

func TestDeductUserBalance(t *testing.T) {
	user := createRandomUser(t)

	// Add initial balance
	initialBalance := float64(200)
	addBalanceArg := AddUserBalanceParams{
		ID:      user.ID,
		Balance: initialBalance,
	}
	_, err := testQueries.AddUserBalance(context.Background(), addBalanceArg)
	require.NoError(t, err)

	deductAmount := float64(100)
	arg := DeductUserBalanceParams{
		ID:      user.ID,
		Balance: deductAmount,
	}
	affectedRows, err := testQueries.DeductUserBalance(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), affectedRows)

	userBalance, err := testQueries.UserBalance(context.Background(), user.ID)
	require.NoError(t, err)
	require.Equal(t, initialBalance-deductAmount, userBalance.Balance)

	// Test insufficient balance
	arg = DeductUserBalanceParams{
		ID:      user.ID,
		Balance: initialBalance,
	}
	affectedRows, err = testQueries.DeductUserBalance(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(0), affectedRows)
}

func TestCheckUserExists(t *testing.T) {
	user := createRandomUser(t)

	exists, err := testQueries.CheckUserExists(context.Background(), user.ID)
	require.NoError(t, err)
	require.True(t, exists)

	nonExistentID := pgtype.UUID{}
	nonExistentID.Bytes = uuid.New()
	exists, err = testQueries.CheckUserExists(context.Background(), nonExistentID)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestCountUsers(t *testing.T) {
	initialCount, err := testQueries.CountUsers(context.Background())
	require.NoError(t, err)

	createRandomUser(t)
	createRandomUser(t)

	newCount, err := testQueries.CountUsers(context.Background())
	require.NoError(t, err)
	require.Equal(t, initialCount+2, newCount)
}

func TestGetUserByUsernameForLogin(t *testing.T) {
	user := createRandomUser(t)

	userByUsername, err := testQueries.GetUserByUsernameForLogin(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, userByUsername)

	require.Equal(t, user.ID, userByUsername.ID)
	require.Equal(t, user.Username, userByUsername.Username)
	require.Equal(t, user.Password, userByUsername.Password)
	require.Equal(t, user.AffiliateID, userByUsername.AffiliateID)
	require.Equal(t, user.Balance, userByUsername.Balance)
}

func TestGetUserDetailByID(t *testing.T) {
	user := createRandomUser(t)

	userDetail, err := testQueries.GetUserDetailByID(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, userDetail)

	require.Equal(t, user.ID, userDetail.ID)
	require.Equal(t, user.Username, userDetail.Username)
	require.Equal(t, user.Balance, userDetail.Balance)
	require.Equal(t, user.AffiliateID, userDetail.AffiliateID)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}

	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}

	users, err := testQueries.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, users, 5)

	for _, user := range users {
		require.NotEmpty(t, user)
	}
}

func TestUserBalance(t *testing.T) {
	user := createRandomUser(t)

	userBalance, err := testQueries.UserBalance(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, userBalance)

	require.Equal(t, user.ID, userBalance.ID)
	require.Equal(t, user.Balance, userBalance.Balance)
}
