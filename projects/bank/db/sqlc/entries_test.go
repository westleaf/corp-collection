package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/westleaf/corp-collection/projects/simplebank/util"
)

func TestCreateEntries(t *testing.T) {
	account1 := createRandomAccount(t)

	arg := CreateEntriesParams{
		AccountID: account1.ID,
		Amount:    util.RandomMoney(),
	}

	_, err := testQueries.CreateEntries(context.Background(), arg)

	require.NoError(t, err)
}

func TestGetEntries(t *testing.T) {
	account1 := createRandomAccount(t)

	arg := CreateEntriesParams{
		AccountID: account1.ID,
		Amount:    util.RandomMoney(),
	}

	entry1, err := testQueries.CreateEntries(context.Background(), arg)
	require.NoError(t, err)

	entry2, err := testQueries.GetEntries(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.Equal(t, entry1, entry2)
}

func TestListEntries(t *testing.T) {
	account1 := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		testQueries.CreateEntries(context.Background(), CreateEntriesParams{
			AccountID: account1.ID,
			Amount:    util.RandomMoney(),
		})
	}

	arg := ListEntriesParams{
		AccountID: account1.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}

}
