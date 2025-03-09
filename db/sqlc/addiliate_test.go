package db

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomAffiliate(t *testing.T) Affiliate {
	masterAffiliate := pgtype.UUID{}
	masterAffiliate.Bytes = uuid.New()
	masterAffiliate.Valid = true

	arg := CreateAffiliateParams{
		Name:            uuid.New().String(),
		MasterAffiliate: masterAffiliate,
	}

	affiliate, err := testQueries.CreateAffiliate(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, affiliate)

	require.Equal(t, arg.Name, affiliate.Name)
	require.Equal(t, arg.MasterAffiliate, affiliate.MasterAffiliate)
	require.Zero(t, affiliate.Balance)
	require.NotZero(t, affiliate.ID)

	return affiliate
}

func TestCreateAffiliate(t *testing.T) {
	createRandomAffiliate(t)
}

func TestAddAffiliateBalance(t *testing.T) {
	affiliate := createRandomAffiliate(t)

	arg := AddAffiliateBalanceParams{
		ID:      affiliate.ID,
		Balance: 100,
	}

	err := testQueries.AddAffiliateBalance(context.Background(), arg)
	require.NoError(t, err)

	affiliateResult, err := testQueries.GetAffiliateByID(context.Background(), affiliate.ID)
	require.NoError(t, err)
	require.Equal(t, arg.Balance, affiliateResult.Balance)
}

func TestGetAffiliateByID(t *testing.T) {
	affiliate1 := createRandomAffiliate(t)
	affiliate2, err := testQueries.GetAffiliateByID(context.Background(), affiliate1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, affiliate2)

	require.Equal(t, affiliate1.ID, affiliate2.ID)
	require.Equal(t, affiliate1.Name, affiliate2.Name)
	require.Equal(t, affiliate1.MasterAffiliate, affiliate2.MasterAffiliate)
	require.Equal(t, affiliate1.Balance, affiliate2.Balance)
}

func TestGetAffiliateByUserID(t *testing.T) {
	affiliate1 := createRandomAffiliate(t)
	affiliate2, err := testQueries.GetAffiliateByUserID(context.Background(), affiliate1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, affiliate2)

	require.Equal(t, affiliate1.ID, affiliate2.ID)
	require.Equal(t, affiliate1.MasterAffiliate, affiliate2.MasterAffiliate)
}

func TestListAffiliates(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAffiliate(t)
	}

	affiliates, err := testQueries.ListAffiliates(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, affiliates)

	for _, affiliate := range affiliates {
		require.NotEmpty(t, affiliate)
	}
}
