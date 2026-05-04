//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/internal/api"
	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

func newTestServer(t *testing.T) (*httptest.Server, *db.DB) {
	t.Helper()
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	h := api.NewHandler(d, func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error) {
		// Stub runner: pretend one new signal happened.
		return []uuid.UUID{uuid.New()}, nil
	})
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)
	return srv, d
}

func doJSON(t *testing.T, method, url string, body any, headers map[string]string) *http.Response {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(buf)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, rdr)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestRegisterDevice(t *testing.T) {
	srv, _ := newTestServer(t)

	resp := doJSON(t, "POST", srv.URL+"/v1/devices", map[string]string{"apns_token": "tok-1"}, nil)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	var out api.RegisterDeviceResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	_, err := uuid.Parse(out.DeviceID)
	require.NoError(t, err)

	// Idempotent.
	resp2 := doJSON(t, "POST", srv.URL+"/v1/devices", map[string]string{"apns_token": "tok-1"}, nil)
	defer resp2.Body.Close()
	var out2 api.RegisterDeviceResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&out2))
	require.Equal(t, out.DeviceID, out2.DeviceID)
}

func TestSubscriptions_Lifecycle(t *testing.T) {
	srv, _ := newTestServer(t)

	// register device
	r := doJSON(t, "POST", srv.URL+"/v1/devices", map[string]string{"apns_token": "tok-1"}, nil)
	defer r.Body.Close()
	var dev api.RegisterDeviceResponse
	json.NewDecoder(r.Body).Decode(&dev)

	hdr := map[string]string{"X-Device-Id": dev.DeviceID}

	// missing X-Device-Id => 401
	resp := doJSON(t, "POST", srv.URL+"/v1/subscriptions",
		api.CreateSubscriptionRequest{Query: "a query", Type: "event", CadenceSeconds: 3600}, nil)
	resp.Body.Close()
	require.Equal(t, 401, resp.StatusCode)

	// validation: cadence too low
	resp = doJSON(t, "POST", srv.URL+"/v1/subscriptions",
		api.CreateSubscriptionRequest{Query: "a query", Type: "event", CadenceSeconds: 60}, hdr)
	resp.Body.Close()
	require.Equal(t, 400, resp.StatusCode)

	// happy path create
	resp = doJSON(t, "POST", srv.URL+"/v1/subscriptions",
		api.CreateSubscriptionRequest{Query: "blockchain events", Type: "event", CadenceSeconds: 3600}, hdr)
	require.Equal(t, 200, resp.StatusCode)
	var sub api.SubscriptionDTO
	json.NewDecoder(resp.Body).Decode(&sub)
	resp.Body.Close()
	require.NotEmpty(t, sub.ID)

	// list
	resp = doJSON(t, "GET", srv.URL+"/v1/subscriptions", nil, hdr)
	require.Equal(t, 200, resp.StatusCode)
	var list []api.SubscriptionDTO
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()
	require.Len(t, list, 1)

	// run (stub)
	resp = doJSON(t, "POST", srv.URL+"/v1/subscriptions/"+sub.ID+"/run", nil, hdr)
	require.Equal(t, 200, resp.StatusCode)
	var run api.RunResponse
	json.NewDecoder(resp.Body).Decode(&run)
	resp.Body.Close()
	require.Equal(t, 1, run.NewSignals)

	// signals (empty since runner is stubbed and didn't actually persist)
	resp = doJSON(t, "GET", srv.URL+"/v1/subscriptions/"+sub.ID+"/signals", nil, hdr)
	require.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()

	// delete
	resp = doJSON(t, "DELETE", srv.URL+"/v1/subscriptions/"+sub.ID, nil, hdr)
	require.Equal(t, 204, resp.StatusCode)
	resp.Body.Close()

	// list again => empty
	resp = doJSON(t, "GET", srv.URL+"/v1/subscriptions", nil, hdr)
	require.Equal(t, 200, resp.StatusCode)
	body, _ := readAll(resp)
	resp.Body.Close()
	require.True(t, strings.Contains(body, "[]"), "expected empty list, got %s", body)
}

func TestCrossDevice_AccessForbidden(t *testing.T) {
	srv, d := newTestServer(t)
	ctx := context.Background()

	// create two devices directly via DB
	devA, _ := d.UpsertDevice(ctx, "tok-A")
	devB, _ := d.UpsertDevice(ctx, "tok-B")

	subA, _ := d.InsertSubscription(ctx, db.SubscriptionInsert{
		DeviceID: devA, Query: "x", Type: "event", CadenceSeconds: 3600,
	})

	// device B tries to delete sub of device A
	hdr := map[string]string{"X-Device-Id": devB.String()}
	resp := doJSON(t, "DELETE", srv.URL+"/v1/subscriptions/"+subA.ID.String(), nil, hdr)
	resp.Body.Close()
	require.Equal(t, 403, resp.StatusCode)
}

func readAll(resp *http.Response) (string, error) {
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}
