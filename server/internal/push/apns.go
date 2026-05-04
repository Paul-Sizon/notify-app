package push

import (
	"context"
	"errors"
	"fmt"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

type APNsConfig struct {
	KeyPath    string
	KeyID      string
	TeamID     string
	BundleID   string
	Production bool
}

type APNsPusher struct {
	cfg    APNsConfig
	client *apns2.Client
}

func NewAPNsPusher(cfg APNsConfig) (*APNsPusher, error) {
	if cfg.KeyPath == "" || cfg.KeyID == "" || cfg.TeamID == "" || cfg.BundleID == "" {
		return nil, errors.New("apns: missing config (KeyPath/KeyID/TeamID/BundleID)")
	}
	authKey, err := token.AuthKeyFromFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("apns: load key: %w", err)
	}
	tok := &token.Token{AuthKey: authKey, KeyID: cfg.KeyID, TeamID: cfg.TeamID}
	c := apns2.NewTokenClient(tok)
	if cfg.Production {
		c = c.Production()
	} else {
		c = c.Development()
	}
	return &APNsPusher{cfg: cfg, client: c}, nil
}

func (p *APNsPusher) Send(ctx context.Context, n Notification) error {
	pl := payload.NewPayload().
		AlertTitle(n.Title).
		AlertBody(n.Body).
		Sound("default").
		Badge(1).
		ThreadID(n.SubscriptionID.String()).
		MutableContent().
		Custom("subscription_id", n.SubscriptionID.String()).
		Custom("signal_id", n.SignalID.String()).
		Custom("url", n.URL)

	notif := &apns2.Notification{
		DeviceToken: n.APNsToken,
		Topic:       p.cfg.BundleID,
		Payload:     pl,
	}
	res, err := p.client.PushWithContext(ctx, notif)
	if err != nil {
		return fmt.Errorf("apns push: %w", err)
	}
	if !res.Sent() {
		return fmt.Errorf("apns rejected: %d %s", res.StatusCode, res.Reason)
	}
	return nil
}
