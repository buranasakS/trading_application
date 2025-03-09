package db

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomProduct(t *testing.T) Product {
	arg := CreateProductParams{
		Name:     uuid.New().String(),
		Quantity: 100,
		Price:    50.0,
	}

	product, err := testQueries.CreateProduct(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, product)

	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.Quantity, product.Quantity)
	require.Equal(t, arg.Price, product.Price)

	require.NotZero(t, product.ID)

	return product
}

func TestCreateProduct(t *testing.T) {
	createRandomProduct(t)
}
func TestGetProductByID(t *testing.T) {
	product1 := createRandomProduct(t)
	product2, err := testQueries.GetProductByID(context.Background(), product1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Quantity, product2.Quantity)
	require.Equal(t, product1.Price, product2.Price)
}

func TestDeductProductQuantity(t *testing.T) {
	product := createRandomProduct(t)

	// Test successful deduction
	deductQuantity := int32(50)
	arg := DeductProductQuantityParams{
		ID:       product.ID,
		Quantity: deductQuantity,
	}
	affectedRows, err := testQueries.DeductProductQuantity(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), affectedRows)

	updatedProduct, err := testQueries.GetProductByID(context.Background(), product.ID)
	require.NoError(t, err)
	require.Equal(t, product.Quantity-deductQuantity, updatedProduct.Quantity)

	arg = DeductProductQuantityParams{
		ID:       product.ID,
		Quantity: updatedProduct.Quantity + 1,
	}
	affectedRows, err = testQueries.DeductProductQuantity(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(0), affectedRows)

	arg = DeductProductQuantityParams{
		ID:       product.ID,
		Quantity: updatedProduct.Quantity,
	}
	affectedRows, err = testQueries.DeductProductQuantity(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, int64(1), affectedRows)
}

func TestListProducts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomProduct(t)
	}

	products, err := testQueries.ListProducts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, products)

	for _, product := range products {
		require.NotEmpty(t, product)
	}
}
