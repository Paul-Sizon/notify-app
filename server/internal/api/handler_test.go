package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCreateSubscription(t *testing.T) {
	cases := []struct {
		name   string
		req    CreateSubscriptionRequest
		wantOK bool
	}{
		{"ok", CreateSubscriptionRequest{Query: "blockchain events", Type: "event", CadenceSeconds: 3600}, true},
		{"short query", CreateSubscriptionRequest{Query: "ab", Type: "event", CadenceSeconds: 3600}, false},
		{"long query", CreateSubscriptionRequest{Query: stringOfLen(201), Type: "event", CadenceSeconds: 3600}, false},
		{"bad type", CreateSubscriptionRequest{Query: "valid query", Type: "podcast", CadenceSeconds: 3600}, false},
		{"low cadence", CreateSubscriptionRequest{Query: "valid query", Type: "event", CadenceSeconds: 60}, false},
		{"news ok", CreateSubscriptionRequest{Query: "valid", Type: "news", CadenceSeconds: 300}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateCreateSubscription(c.req)
			if c.wantOK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
