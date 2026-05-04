//go:build integration

package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/paulsizon/notify/server/e2e"
	"github.com/paulsizon/notify/server/internal/agent"
	"github.com/paulsizon/notify/server/internal/api"
	"github.com/paulsizon/notify/server/internal/db"
	"github.com/paulsizon/notify/server/internal/push"
	"github.com/paulsizon/notify/server/internal/testhelpers"
)

// scriptedExtractor lets each test pre-load the events / news that the stub
// pipeline will emit. Per-query lookup falls back to defaults.
type scriptedExtractor struct {
	mu             sync.Mutex
	eventsByQuery  map[string][]agent.EventCandidate
	newsByQuery    map[string]agent.NewsExtraction
	defaultEvents  []agent.EventCandidate
	defaultNews    agent.NewsExtraction
	extractCalls   int
}

func newScriptedExtractor() *scriptedExtractor {
	return &scriptedExtractor{
		eventsByQuery: map[string][]agent.EventCandidate{},
		newsByQuery:   map[string]agent.NewsExtraction{},
	}
}

func (s *scriptedExtractor) ExtractEvents(ctx context.Context, in agent.ExtractInput) ([]agent.EventCandidate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extractCalls++
	if v, ok := s.eventsByQuery[in.Query]; ok {
		return v, nil
	}
	return s.defaultEvents, nil
}

func (s *scriptedExtractor) ExtractNews(ctx context.Context, in agent.ExtractInput) (agent.NewsExtraction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extractCalls++
	if v, ok := s.newsByQuery[in.Query]; ok {
		return v, nil
	}
	return s.defaultNews, nil
}

// pushSpy records every send so tests can assert notification side-effects.
type pushSpy struct {
	mu   sync.Mutex
	sent []push.Notification
}

func (p *pushSpy) Send(ctx context.Context, n push.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = append(p.sent, n)
	return nil
}
func (p *pushSpy) Sent() []push.Notification {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]push.Notification, len(p.sent))
	copy(out, p.sent)
	return out
}

type e2eHarness struct {
	URL       string
	Server    *httptest.Server
	DB        *db.DB
	Extractor *scriptedExtractor
	Pusher    *pushSpy
}

func newE2EHarness(t *testing.T) *e2eHarness {
	t.Helper()
	pool := testhelpers.TestDBPool(t)
	d := db.New(pool)
	ext := newScriptedExtractor()
	spy := &pushSpy{}

	deps := agent.Deps{
		DB:        d,
		Searcher:  stubSearcher{},
		Extractor: ext,
		Pusher:    spy,
	}
	runner := func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error) {
		return agent.RunSubscription(ctx, deps, subID)
	}
	h := api.NewHandler(d, runner)
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)
	return &e2eHarness{URL: srv.URL, Server: srv, DB: d, Extractor: ext, Pusher: spy}
}

func makeEvent(title, date, venue, city string) agent.EventCandidate {
	return agent.EventCandidate{
		Title:      title,
		Date:       sP(date),
		Venue:      sP(venue),
		City:       sP(city),
		Confidence: 0.9,
	}
}

// ---------------------------------------------------------------------------
//
//                          User-flow E2E tests
//
// ---------------------------------------------------------------------------

// First-launch flow: app registers, has no subscriptions, list returns [].
func TestUser_FirstLaunch_EmptyList(t *testing.T) {
	h := newE2EHarness(t)
	c := e2e.NewClient(h.URL)

	devID, err := c.RegisterDevice("first-launch-tok")
	require.NoError(t, err)
	require.NotEmpty(t, devID)

	subs, err := c.ListSubscriptions()
	require.NoError(t, err)
	require.Empty(t, subs, "fresh device must see empty subscription list")
}

// Reinstall / app-restart flow: re-register identical APNs token, expect same
// device_id, see prior subscriptions intact. Mirrors what iOS does on every
// cold launch.
func TestUser_AppRestart_PreservesSubscriptions(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.defaultEvents = []agent.EventCandidate{
		makeEvent("Coldplay", "2026-09-05", "Allianz Parque", "São Paulo"),
	}

	// Session 1.
	c1 := e2e.NewClient(h.URL)
	devID1, err := c1.RegisterDevice("persistent-tok")
	require.NoError(t, err)
	sub1, err := c1.CreateSubscription("Coldplay 2026", "event", 3600)
	require.NoError(t, err)
	_, err = c1.RunSubscription(sub1.ID)
	require.NoError(t, err)

	// Simulate app kill + relaunch — fresh client, same APNs token.
	c2 := e2e.NewClient(h.URL)
	devID2, err := c2.RegisterDevice("persistent-tok")
	require.NoError(t, err)
	require.Equal(t, devID1, devID2, "same token => same device_id across launches")

	subs, err := c2.ListSubscriptions()
	require.NoError(t, err)
	require.Len(t, subs, 1)
	require.Equal(t, sub1.ID, subs[0].ID)

	sigs, err := c2.ListSignals(sub1.ID)
	require.NoError(t, err)
	require.Len(t, sigs, 1, "signals from prior session still visible")
}

// Multiple subscriptions per device, each independent. Pull-to-refresh on
// subscription A must not affect B's signal count.
func TestUser_MultipleSubscriptions_Independent(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.eventsByQuery["Coldplay 2026"] = []agent.EventCandidate{
		makeEvent("Coldplay SP", "2026-09-05", "Allianz Parque", "São Paulo"),
		makeEvent("Coldplay RJ", "2026-09-08", "Nilton Santos", "Rio de Janeiro"),
	}
	h.Extractor.eventsByQuery["Arctic Monkeys 2026"] = []agent.EventCandidate{
		makeEvent("Arctic Monkeys SP", "2026-11-10", "Allianz Parque", "São Paulo"),
	}
	h.Extractor.eventsByQuery["Tame Impala 2026"] = []agent.EventCandidate{
		makeEvent("Tame Impala SP", "2026-12-01", "Espaço Unimed", "São Paulo"),
		makeEvent("Tame Impala BSB", "2026-12-04", "Mané Garrincha", "Brasília"),
		makeEvent("Tame Impala BH", "2026-12-07", "Mineirão", "Belo Horizonte"),
	}

	c := e2e.NewClient(h.URL)
	_, err := c.RegisterDevice("multi-sub-tok")
	require.NoError(t, err)

	queries := []struct {
		q       string
		expect  int
	}{
		{"Coldplay 2026", 2},
		{"Arctic Monkeys 2026", 1},
		{"Tame Impala 2026", 3},
	}
	subs := make(map[string]string)
	for _, qc := range queries {
		s, err := c.CreateSubscription(qc.q, "event", 3600)
		require.NoError(t, err)
		subs[qc.q] = s.ID

		run, err := c.RunSubscription(s.ID)
		require.NoError(t, err)
		require.Equal(t, qc.expect, run.NewSignals, "query %s", qc.q)
	}

	// Refresh just subscription A — others unaffected.
	_, err = c.RunSubscription(subs["Coldplay 2026"])
	require.NoError(t, err)

	for _, qc := range queries {
		sigs, err := c.ListSignals(subs[qc.q])
		require.NoError(t, err)
		require.Len(t, sigs, qc.expect, "query %s post-refresh-A", qc.q)
	}
}

// Two devices on the same query each maintain their own signal stream. Deleting
// one device's subscription does not affect the other's.
func TestUser_TwoDevices_SameQuery_AreIsolated(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.defaultEvents = []agent.EventCandidate{
		makeEvent("Coldplay", "2026-09-05", "Allianz", "São Paulo"),
	}

	cA := e2e.NewClient(h.URL)
	_, err := cA.RegisterDevice("device-A")
	require.NoError(t, err)
	subA, err := cA.CreateSubscription("Coldplay 2026", "event", 3600)
	require.NoError(t, err)

	cB := e2e.NewClient(h.URL)
	_, err = cB.RegisterDevice("device-B")
	require.NoError(t, err)
	subB, err := cB.CreateSubscription("Coldplay 2026", "event", 3600)
	require.NoError(t, err)
	require.NotEqual(t, subA.ID, subB.ID, "each device gets its own subscription row")

	_, err = cA.RunSubscription(subA.ID)
	require.NoError(t, err)
	_, err = cB.RunSubscription(subB.ID)
	require.NoError(t, err)

	// Each sees their own one signal.
	sA, err := cA.ListSignals(subA.ID)
	require.NoError(t, err)
	require.Len(t, sA, 1)
	sB, err := cB.ListSignals(subB.ID)
	require.NoError(t, err)
	require.Len(t, sB, 1)

	// Device A cannot list device B's signals.
	req, err := http.NewRequest(http.MethodGet, h.URL+"/v1/subscriptions/"+subB.ID+"/signals", nil)
	require.NoError(t, err)
	req.Header.Set("X-Device-Id", devIDFromClient(t, cA))
	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	r.Body.Close()
	require.Equal(t, http.StatusForbidden, r.StatusCode)

	// Device A deletes its sub. Device B's sub still works.
	require.NoError(t, cA.DeleteSubscription(subA.ID))
	sB2, err := cB.ListSignals(subB.ID)
	require.NoError(t, err)
	require.Len(t, sB2, 1)
}

// Pull-to-refresh spam: tap pull-to-refresh 5 times in quick succession. Total
// signals must not increase past the deterministic stub output.
func TestUser_PullToRefreshSpam_NoDuplicates(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.defaultEvents = []agent.EventCandidate{
		makeEvent("Coldplay", "2026-09-05", "Allianz", "São Paulo"),
		makeEvent("Coldplay", "2026-09-08", "Nilton Santos", "Rio de Janeiro"),
	}

	c := e2e.NewClient(h.URL)
	_, err := c.RegisterDevice("spammer")
	require.NoError(t, err)
	sub, err := c.CreateSubscription("Coldplay", "event", 3600)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err := c.RunSubscription(sub.ID)
		require.NoError(t, err)
	}

	sigs, err := c.ListSignals(sub.ID)
	require.NoError(t, err)
	require.Len(t, sigs, 2, "5 refreshes must not multiply signals")

	// Push fired only on the first run (other runs produced 0 new signals).
	require.Len(t, h.Pusher.Sent(), 2, "push only on actually-new signals")
}

// Delete + recreate: user removes a subscription, creates one with the same
// query. New subscription gets a fresh ID and re-runs produce signals from
// scratch (no leftover dedup state from the deleted sub).
func TestUser_DeleteThenRecreate_StartsFresh(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.defaultEvents = []agent.EventCandidate{
		makeEvent("Coldplay", "2026-09-05", "Allianz", "São Paulo"),
	}

	c := e2e.NewClient(h.URL)
	_, err := c.RegisterDevice("recreate-tok")
	require.NoError(t, err)

	sub1, err := c.CreateSubscription("Coldplay", "event", 3600)
	require.NoError(t, err)
	_, err = c.RunSubscription(sub1.ID)
	require.NoError(t, err)
	sigs1, err := c.ListSignals(sub1.ID)
	require.NoError(t, err)
	require.Len(t, sigs1, 1)

	require.NoError(t, c.DeleteSubscription(sub1.ID))

	sub2, err := c.CreateSubscription("Coldplay", "event", 3600)
	require.NoError(t, err)
	require.NotEqual(t, sub1.ID, sub2.ID)

	run, err := c.RunSubscription(sub2.ID)
	require.NoError(t, err)
	require.Equal(t, 1, run.NewSignals, "recreate must yield fresh signal, dedup state was scoped to old sub")

	sigs2, err := c.ListSignals(sub2.ID)
	require.NoError(t, err)
	require.Len(t, sigs2, 1)
	require.NotEqual(t, sigs1[0].ID, sigs2[0].ID)
}

// Pagination: client uses limit param to control page size. Cursor-based
// next-page is exercised across two batches with distinct insert timestamps.
func TestUser_Pagination_LimitAndCursor(t *testing.T) {
	h := newE2EHarness(t)

	c := e2e.NewClient(h.URL)
	_, err := c.RegisterDevice("pager")
	require.NoError(t, err)
	sub, err := c.CreateSubscription("multi", "event", 3600)
	require.NoError(t, err)

	// First batch — 3 signals.
	h.Extractor.defaultEvents = []agent.EventCandidate{
		makeEvent("Show 0", "2026-09-01", "V0", "C0"),
		makeEvent("Show 1", "2026-09-02", "V1", "C1"),
		makeEvent("Show 2", "2026-09-03", "V2", "C2"),
	}
	_, err = c.RunSubscription(sub.ID)
	require.NoError(t, err)

	// Force a measurable gap so cursor-based paging splits cleanly between batches.
	time.Sleep(50 * time.Millisecond)

	// Second batch — 2 more signals at clearly-later first_seen_at.
	h.Extractor.defaultEvents = []agent.EventCandidate{
		makeEvent("Show 3", "2026-09-04", "V3", "C3"),
		makeEvent("Show 4", "2026-09-05", "V4", "C4"),
	}
	_, err = c.RunSubscription(sub.ID)
	require.NoError(t, err)

	// Default list — all 5.
	all, err := c.ListSignals(sub.ID)
	require.NoError(t, err)
	require.Len(t, all, 5)

	// limit=2 — first 2 (newest first; should be from batch 2).
	page1 := mustList(t, h.URL, c, sub.ID, "?limit=2")
	require.Len(t, page1, 2)

	// Use the boundary timestamp between batches as cursor — should yield batch 1.
	cursor := all[2].FirstSeenAt.UTC().Add(time.Millisecond).Format(time.RFC3339Nano)
	page2 := mustList(t, h.URL, c, sub.ID, "?before="+cursor)
	require.Len(t, page2, 3, "older page should contain the 3 batch-1 signals")
}

// Validation flow: the iOS form catches most issues but the server is the
// source of truth. Confirm friendly status codes for typical mistakes.
func TestUser_Validation_RejectsBadInputs(t *testing.T) {
	h := newE2EHarness(t)

	// no device header
	r, err := http.Post(h.URL+"/v1/subscriptions", "application/json",
		newBody(`{"query":"valid","type":"event","cadence_seconds":3600}`))
	require.NoError(t, err)
	r.Body.Close()
	require.Equal(t, http.StatusUnauthorized, r.StatusCode)

	// garbage device header
	req, _ := http.NewRequest("POST", h.URL+"/v1/subscriptions",
		newBody(`{"query":"valid","type":"event","cadence_seconds":3600}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-Id", "not-a-uuid")
	r, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	r.Body.Close()
	require.Equal(t, http.StatusBadRequest, r.StatusCode)

	// real device, then bad payloads
	c := e2e.NewClient(h.URL)
	_, err = c.RegisterDevice("validator")
	require.NoError(t, err)

	cases := []struct {
		name    string
		payload string
		want    int
	}{
		{"short query", `{"query":"ab","type":"event","cadence_seconds":3600}`, 400},
		{"bad type", `{"query":"valid","type":"podcast","cadence_seconds":3600}`, 400},
		{"low cadence", `{"query":"valid","type":"event","cadence_seconds":60}`, 400},
		{"malformed json", `{"query":"valid",}`, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", h.URL+"/v1/subscriptions", newBody(tc.payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Device-Id", devIDFromClient(t, c))
			r, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			r.Body.Close()
			require.Equal(t, tc.want, r.StatusCode)
		})
	}
}

// Tap-notification flow simulation: server returns a signal with payload that
// the iOS app would deep-link to. Confirm subscription_id + signal_id are
// usable to fetch detail.
func TestUser_NotificationDeepLink_SignalRetrievableByID(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.defaultEvents = []agent.EventCandidate{
		{Title: "Coldplay", Date: sP("2026-09-05"), Venue: sP("Allianz"), City: sP("SP"),
			URL: sP("https://example.com/coldplay-sp"), Confidence: 0.95},
	}

	c := e2e.NewClient(h.URL)
	_, err := c.RegisterDevice("deeplink-tok")
	require.NoError(t, err)
	sub, err := c.CreateSubscription("Coldplay", "event", 3600)
	require.NoError(t, err)

	_, err = c.RunSubscription(sub.ID)
	require.NoError(t, err)

	// Push fired and includes both IDs (what APNs payload would carry).
	sent := h.Pusher.Sent()
	require.Len(t, sent, 1)
	require.Equal(t, sub.ID, sent[0].SubscriptionID.String())
	require.NotEqual(t, "", sent[0].SignalID.String())
	require.Equal(t, "https://example.com/coldplay-sp", sent[0].URL)

	// iOS would now hit GET /signals to render detail. Verify the IDs match.
	sigs, err := c.ListSignals(sub.ID)
	require.NoError(t, err)
	require.Len(t, sigs, 1)
	require.Equal(t, sent[0].SignalID.String(), sigs[0].ID)
}

// Concurrent runs of two different subscriptions should not interfere.
func TestUser_ConcurrentRuns_DoNotCorrupt(t *testing.T) {
	h := newE2EHarness(t)
	h.Extractor.eventsByQuery["query alpha"] = []agent.EventCandidate{
		makeEvent("A1", "2026-09-05", "V", "C"),
		makeEvent("A2", "2026-09-06", "V", "C"),
	}
	h.Extractor.eventsByQuery["query bravo"] = []agent.EventCandidate{
		makeEvent("B1", "2026-10-05", "V", "C"),
	}

	c := e2e.NewClient(h.URL)
	_, err := c.RegisterDevice("concurrent")
	require.NoError(t, err)
	subA, err := c.CreateSubscription("query alpha", "event", 3600)
	require.NoError(t, err)
	subB, err := c.CreateSubscription("query bravo", "event", 3600)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = c.RunSubscription(subA.ID) }()
	go func() { defer wg.Done(); _, _ = c.RunSubscription(subB.ID) }()
	wg.Wait()

	sA, err := c.ListSignals(subA.ID)
	require.NoError(t, err)
	require.Len(t, sA, 2)
	sB, err := c.ListSignals(subB.ID)
	require.NoError(t, err)
	require.Len(t, sB, 1)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func devIDFromClient(t *testing.T, c *e2e.Client) string {
	t.Helper()
	require.NotEmpty(t, c.DeviceID)
	return c.DeviceID
}

func mustList(t *testing.T, base string, c *e2e.Client, subID, query string) []api.SignalDTO {
	t.Helper()
	req, _ := http.NewRequest("GET", base+"/v1/subscriptions/"+subID+"/signals"+query, nil)
	req.Header.Set("X-Device-Id", c.DeviceID)
	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer r.Body.Close()
	require.Equal(t, 200, r.StatusCode)
	var out []api.SignalDTO
	require.NoError(t, json.NewDecoder(r.Body).Decode(&out))
	return out
}

func newBody(s string) io.Reader { return strings.NewReader(s) }
