package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	println("bal before:", acc1.Balance, acc2.Balance)

	n := 5
	amount := int64(10)

	errs := make(chan error)
	rests := make(chan TransferTxResult)

	for i := range n {
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			res, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})
			errs <- err
			rests <- res
		}()
	}

	existed := make(map[int]bool)

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

		fromacc := res.FromAccount
		require.NotEmpty(t, fromacc)
		require.Equal(t, acc1.ID, fromacc.ID)

		toacc := res.ToAccount
		require.NotEmpty(t, toacc)
		require.Equal(t, acc2.ID, toacc.ID)

		diff1 := acc1.Balance - fromacc.Balance
		diff2 := toacc.Balance - acc2.Balance
		println("tx:", fromacc.Balance, toacc.Balance)
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updacc1, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	updacc2, err := testQueries.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	println("after before:", updacc1.Balance, updacc2.Balance)

	require.Equal(t, updacc1.Balance, acc1.Balance-int64(n)*amount)
	require.Equal(t, updacc2.Balance, acc2.Balance+int64(n)*amount)

}
