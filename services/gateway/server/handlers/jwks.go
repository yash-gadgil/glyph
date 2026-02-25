package handlers

import (
	"context"
	"crypto/rsa"
	"errors"
	"sync"
	"time"

	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/types/known/emptypb"
)

const jwksTTL = time.Hour

const jwksMinRefresh = 30 * time.Second

var errUnknownKID = errors.New("unknown signing key id")

type jwksCache struct {
	client authpb.AuthServiceClient
	ttl    time.Duration

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time

	group singleflight.Group
}

func newJWKSCache(client authpb.AuthServiceClient) *jwksCache {
	return &jwksCache{
		client: client,
		ttl:    jwksTTL,
		keys:   map[string]*rsa.PublicKey{},
	}
}

func (c *jwksCache) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("jwks cache not configured")
	}

	c.mu.RLock()
	key, ok := c.keys[kid]
	age := time.Since(c.fetchedAt)
	c.mu.RUnlock()

	if ok && age < c.ttl {
		return key, nil
	}

	if !ok && age < jwksMinRefresh {
		return nil, errUnknownKID
	}

	if err := c.refresh(ctx); err != nil {
		if ok {
			return key, nil
		}
		return nil, err
	}

	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok {
		return nil, errUnknownKID
	}
	return key, nil
}

func (c *jwksCache) refresh(ctx context.Context) error {
	_, err, _ := c.group.Do("jwks", func() (any, error) {
		resp, err := c.client.GetPublicKeys(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, err
		}

		keys := make(map[string]*rsa.PublicKey, len(resp.Keys))
		for _, k := range resp.Keys {
			pub, err := jwksToRSAPublicKey(k.N, k.E)
			if err != nil {
				continue
			}
			keys[k.Kid] = pub
		}

		c.mu.Lock()
		c.keys = keys
		c.fetchedAt = time.Now()
		c.mu.Unlock()
		return nil, nil
	})
	return err
}

func (c *jwksCache) expireForTest() {
	c.mu.Lock()
	c.fetchedAt = time.Time{}
	c.mu.Unlock()
}
