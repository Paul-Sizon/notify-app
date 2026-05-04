package agent

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func sP(s string) *string { return &s }

func TestEventsToSignals_BuildsStableFingerprints(t *testing.T) {
	subID := uuid.New()
	cands := []EventCandidate{
		{Title: "Coldplay", Date: sP("2026-09-05"), Venue: sP("Allianz Parque"), City: sP("São Paulo"), Confidence: 0.9},
		{Title: "coldplay", Date: sP("2026-09-05"), Venue: sP("allianz parque"), City: sP("são paulo"), Confidence: 0.9},
	}
	sigs := eventsToSignals(subID, cands)
	require.Len(t, sigs, 2)
	require.Equal(t, sigs[0].Fingerprint, sigs[1].Fingerprint, "casing should not change fp")
}

func TestEventsToSignals_ParsesOccursAt(t *testing.T) {
	subID := uuid.New()
	cands := []EventCandidate{
		{Title: "X", Date: sP("2026-09-05"), Venue: sP("V"), Confidence: 0.9},
	}
	sigs := eventsToSignals(subID, cands)
	require.Len(t, sigs, 1)
	require.NotNil(t, sigs[0].OccursAt)
	require.Equal(t, 2026, sigs[0].OccursAt.Year())
}

func TestNewsToSignals_DropsNonNewDevelopments(t *testing.T) {
	subID := uuid.New()
	items := []NewsCandidate{
		{Headline: "A", CanonicalHeadline: "a", IsNewDevelopment: true, Confidence: 0.9},
		{Headline: "B", CanonicalHeadline: "b", IsNewDevelopment: false, Confidence: 0.9},
	}
	sigs := newsToSignals(subID, items)
	require.Len(t, sigs, 1)
	require.Equal(t, "A", sigs[0].Title)
}

func TestDomainsFromURL(t *testing.T) {
	require.Equal(t, []string{"example.com"}, domainsFromURL(sP("https://www.example.com/page")))
	require.Empty(t, domainsFromURL(nil))
	require.Empty(t, domainsFromURL(sP("")))
}
