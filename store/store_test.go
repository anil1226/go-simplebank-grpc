package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	n := 5
	amount := int64(10)

	errs := make(chan error)
	rests := make(chan TransferTxResult)

	for range n {
		go func() {
			res, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})
			errs <- err
			rests <- res
		}()
	}

	for range n {
		err := <-errs
		require.NoError(t, err)

		res := <-rests
		require.NotEmpty(t, res)

		trans := res.Transfer

		require.NotEmpty(t, trans)
		require.Equal(t, acc1.ID, trans.FromAccountID)
		require.Equal(t, acc2.ID, trans.ToAccountID)
		require.Equal(t, amount, trans.Amount)
		require.NotZero(t, trans.ID)
		require.NotZero(t, trans.CreatedAt)
	}
}
