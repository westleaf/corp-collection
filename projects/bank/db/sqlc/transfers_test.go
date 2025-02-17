package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/westleaf/corp-collection/projects/simplebank/util"
)

func createRandomTransfer(t *testing.T) Transfer {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	transfer, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomInt(0, account1.Balance),
	})
	require.NoError(t, err)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	arg := UpdateAccountParams{
		Balance: account1.Balance + account2.Balance,
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        account1.Balance,
	})

	require.NoError(t, err)
	require.Equal(t, transfer.Amount, account1.Balance)
	require.Equal(t, arg.Balance, (transfer.Amount + account2.Balance))
}

func TestGetTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	transfer1, err := testQueries.CreateTransfer(context.Background(), CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        account1.Balance,
	})

	transfer2, err := testQueries.GetTrasfer(context.Background(), transfer1.ID)

	require.NoError(t, err)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.Equal(t, transfer1.CreatedAt, transfer2.CreatedAt)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
}

func TestListTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	params := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomInt(1, account1.Balance/5),
	}

	for i := 0; i < 5; i++ {
		transfer, err := testQueries.CreateTransfer(context.Background(), params)
		require.NoError(t, err)
		require.NotEmpty(t, transfer)
	}

	arg := ListTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        0,
	}

	transfers, err := testQueries.ListTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
