package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	rsaKeySize              = 2048
	defaultRotationInterval = 30 * 24 * time.Hour
	keyGracePeriod          = 7 * 24 * time.Hour
)

type KeyPair struct {
	ID         string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type KeyStore struct {
	currentKey *KeyPair
	oldKeys    map[string]*KeyPair
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewKeyStore(logger *zap.Logger) (*KeyStore, error) {
	ks := &KeyStore{
		oldKeys: make(map[string]*KeyPair),
		logger:  logger,
	}

	if err := ks.rotateKey(); err != nil {
		return nil, fmt.Errorf("failed to generate initial key pair: %w", err)
	}

	ks.logger.Info("keystore_initialized", zap.String("key_id", ks.currentKey.ID))
	return ks, nil
}

func (ks *KeyStore) GetCurrentKey() *KeyPair {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return ks.currentKey
}

func (ks *KeyStore) GetPublicKey(kid string) (*rsa.PublicKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	if ks.currentKey != nil && ks.currentKey.ID == kid {
		return ks.currentKey.PublicKey, nil
	}

	if key, ok := ks.oldKeys[kid]; ok {
		return key.PublicKey, nil
	}

	return nil, fmt.Errorf("public key not found for kid: %s", kid)
}

func (ks *KeyStore) GetAllPublicKeys() []*KeyPair {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keys := make([]*KeyPair, 0, len(ks.oldKeys)+1)

	if ks.currentKey != nil {
		keys = append(keys, ks.currentKey)
	}

	now := time.Now()
	for _, key := range ks.oldKeys {
		if now.Before(key.ExpiresAt) {
			keys = append(keys, key)
		}
	}

	return keys
}

func (ks *KeyStore) rotateKey() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return fmt.Errorf("failed to generate key id suffix: %w", err)
	}
	kid := fmt.Sprintf("key-%d-%x", time.Now().Unix(), suffix)

	newKey := &KeyPair{
		ID:         kid,
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(defaultRotationInterval + keyGracePeriod),
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.currentKey != nil {
		ks.oldKeys[ks.currentKey.ID] = ks.currentKey
	}

	ks.currentKey = newKey
	ks.cleanupExpiredKeys()

	ks.logger.Info("key_rotation_completed",
		zap.String("new_key_id", kid),
		zap.Int("active_old_keys", len(ks.oldKeys)))

	return nil
}

func (ks *KeyStore) cleanupExpiredKeys() {
	now := time.Now()
	for kid, key := range ks.oldKeys {
		if now.After(key.ExpiresAt) {
			delete(ks.oldKeys, kid)
		}
	}
}

func (ks *KeyStore) StartRotationScheduler(interval time.Duration) chan struct{} {
	if interval == 0 {
		interval = defaultRotationInterval
	}

	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ks.rotateKey(); err != nil {
					ks.logger.Error("key_rotation_failed", zap.Error(err))
				}
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

func ExportPublicKeyPEM(pub *rsa.PublicKey) (string, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return string(pubPEM), nil
}

func PublicKeyToJWK(kid string, pub *rsa.PublicKey) map[string]string {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	return map[string]string{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": kid,
		"n":   n,
		"e":   e,
	}
}
