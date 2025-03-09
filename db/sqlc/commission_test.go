package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomCommission(t *testing.T) Commission {
	affiliate := createRandomAffiliate(t)

	arg := CreateCommissionParams{
		OrderID:     pgtype.UUID{Bytes: uuid.New(), Valid: true},
		AffiliateID: affiliate.ID,
		Amount:      100.50,
	}

	commission, err := testQueries.CreateCommission(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, commission)

	require.Equal(t, arg.OrderID, commission.OrderID)
	require.Equal(t, arg.AffiliateID, commission.AffiliateID)
	require.Equal(t, arg.Amount, commission.Amount)

	require.NotZero(t, commission.ID)

	return commission
}

func TestCreateCommission(t *testing.T) {
	createRandomCommission(t)
}

func TestGetCommissionByID(t *testing.T) {
	commission1 := createRandomCommission(t)
	commission2, err := testQueries.GetCommissionByID(context.Background(), commission1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, commission2)

	require.Equal(t, commission1.ID, commission2.ID)
	require.Equal(t, commission1.OrderID, commission2.OrderID)
	require.Equal(t, commission1.AffiliateID, commission2.AffiliateID)
	require.Equal(t, commission1.Amount, commission2.Amount)
}

func TestListCommissions(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomCommission(t)
	}

	commissions, err := testQueries.ListCommissions(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, commissions)

	for _, commission := range commissions {
		require.NotEmpty(t, commission)
	}
}
