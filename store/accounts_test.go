package store

import (
	"context"
	"testing"

	"github.com/anil1226/go-simplebank-grpc/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	return acc
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)

	require.Equal(t, acc1, acc2)
}

func TestUpdateAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	args := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: util.RandomMoney(),
	}

	_, err := testQueries.UpdateAccount(context.Background(), args)
	require.NoError(t, err)
}

func TestDeleteAccount(t *testing.T) {
	acc1 := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
}

func TestListAccount(t *testing.T) {
	var lastAccount Account
	for range 3 {
		lastAccount = createRandomAccount(t)
	}

	args := ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	acc, err := testQueries.ListAccounts(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, acc)

	for _, ac := range acc {
		require.NotEmpty(t, ac)
		require.Equal(t, lastAccount.Owner, ac.Owner)
	}
}
