package agent

import "testing"

func TestEventFingerprint_StableAcrossWhitespaceAndCase(t *testing.T) {
	a := EventFingerprint("Taylor Swift  LIVE!", "2025-08-12", "Allianz Parque")
	b := EventFingerprint("taylor swift live", "2025-08-12", "allianz parque")
	if a != b {
		t.Fatalf("expected equal, got\n%s\n%s", a, b)
	}
}

func TestEventFingerprint_DifferentDateDifferent(t *testing.T) {
	a := EventFingerprint("Taylor Swift", "2025-08-12", "Allianz")
	b := EventFingerprint("Taylor Swift", "2025-08-13", "Allianz")
	if a == b {
		t.Fatal("expected different fingerprint when date differs")
	}
}

func TestEventFingerprint_StripsStopwords(t *testing.T) {
	a := EventFingerprint("The Cure Live Tour", "2025-09-01", "Ginásio")
	b := EventFingerprint("Cure", "2025-09-01", "Ginásio")
	if a != b {
		t.Fatalf("stopwords not stripped:\n%s\n%s", a, b)
	}
}

func TestEventFingerprint_EmptyVenueDifferentFromNonEmpty(t *testing.T) {
	a := EventFingerprint("Taylor Swift", "2025-08-12", "")
	b := EventFingerprint("Taylor Swift", "2025-08-12", "Allianz")
	if a == b {
		t.Fatal("empty venue should differ from non-empty")
	}
}

func TestNewsFingerprint_NormalizesPunctAndCase(t *testing.T) {
	a := NewsFingerprint("SEC Approves Bitcoin ETF")
	b := NewsFingerprint("sec approves bitcoin etf!!!")
	if a != b {
		t.Fatalf("expected equal:\n%s\n%s", a, b)
	}
}

func TestNewsFingerprint_DifferentHeadlinesDiffer(t *testing.T) {
	a := NewsFingerprint("SEC approves Bitcoin ETF")
	b := NewsFingerprint("SEC rejects Bitcoin ETF")
	if a == b {
		t.Fatal("expected different fingerprints")
	}
}

func TestNormalize_CollapsesWhitespace(t *testing.T) {
	got := normalize("  Hello\t\nWorld  ")
	want := "hello world"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalize_UnicodeLetters(t *testing.T) {
	got := normalize("Ginásio do Atlético")
	want := "ginásio do atlético"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
