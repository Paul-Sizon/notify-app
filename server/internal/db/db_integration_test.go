//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

func ctxT(t *testing.T) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func TestUpsertDevice_IsIdempotentByToken(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx, cancel := ctxT(t)
	defer cancel()

	id1, err := d.UpsertDevice(ctx, "tok-abc")
	require.NoError(t, err)
	id2, err := d.UpsertDevice(ctx, "tok-abc")
	require.NoError(t, err)
	require.Equal(t, id1, id2, "same token should return same id")

	id3, err := d.UpsertDevice(ctx, "tok-xyz")
	require.NoError(t, err)
	require.NotEqual(t, id1, id3)
}

func TestSubscription_CRUD(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx, cancel := ctxT(t)
	defer cancel()

	devID, _ := d.UpsertDevice(ctx, "tok-1")

	sub, err := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "blockchain events curitiba", Type: "event", CadenceSeconds: 3600,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, sub.ID)

	got, err := d.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, sub.Query, got.Query)

	list, err := d.ListSubscriptionsByDevice(ctx, devID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, d.DeleteSubscription(ctx, sub.ID))
	list, err = d.ListSubscriptionsByDevice(ctx, devID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestInsertSignals_OnConflictReturnsOnlyNew(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx, cancel := ctxT(t)
	defer cancel()

	devID, _ := d.UpsertDevice(ctx, "tok-1")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "x", Type: "event", CadenceSeconds: 3600,
	})

	makeSig := func(fp, title string) db.SignalInsert {
		return db.SignalInsert{
			SubscriptionID: sub.ID,
			Fingerprint:    fp,
			Title:          title,
			Confidence:     0.9,
			Payload:        []byte(`{}`),
			SourceDomains:  []string{},
		}
	}

	ids, err := d.InsertSignals(ctx, []db.SignalInsert{
		makeSig("fp-1", "A"),
		makeSig("fp-2", "B"),
	})
	require.NoError(t, err)
	require.Len(t, ids, 2, "first insert: both new")

	ids2, err := d.InsertSignals(ctx, []db.SignalInsert{
		makeSig("fp-2", "B'"),
		makeSig("fp-3", "C"),
	})
	require.NoError(t, err)
	require.Len(t, ids2, 1, "fp-2 dedup, fp-3 new")
}

func TestListDueSubscriptions_OrdersAndFilters(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx, cancel := ctxT(t)
	defer cancel()

	devID, _ := d.UpsertDevice(ctx, "tok-1")
	subA, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "a", Type: "event", CadenceSeconds: 3600,
	})
	subB, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "b", Type: "event", CadenceSeconds: 3600,
	})

	// Push subB into the future explicitly.
	future := time.Now().Add(10 * time.Minute)
	_, err := pool.Exec(ctx, `UPDATE subscriptions SET next_run_at = $2 WHERE id = $1`, subB.ID, future)
	require.NoError(t, err)

	due, err := d.ListDueSubscriptions(ctx, time.Now(), 10)
	require.NoError(t, err)
	require.Len(t, due, 1)
	require.Equal(t, subA.ID, due[0].ID)
}

func TestRescheduleSubscription_AdvancesNextRun(t *testing.T) {
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ctx, cancel := ctxT(t)
	defer cancel()

	devID, _ := d.UpsertDevice(ctx, "tok-1")
	sub, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devID, Query: "a", Type: "event", CadenceSeconds: 3600,
	})

	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, d.RescheduleSubscription(ctx, sub.ID, now, "summary v2"))

	got, err := d.GetSubscription(ctx, sub.ID)
	require.NoError(t, err)
	require.NotNil(t, got.LastRunAt)
	require.WithinDuration(t, now, *got.LastRunAt, time.Second)
	require.WithinDuration(t, now.Add(time.Hour), got.NextRunAt, time.Second)
	require.Equal(t, "summary v2", got.RollingSummary)
}
