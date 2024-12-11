package auth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	grpc "github.com/Layr-Labs/eigenda/api/grpc/node/v2"
	"github.com/Layr-Labs/eigenda/core"
	dauth "github.com/Layr-Labs/eigenda/disperser/auth"
	"github.com/ethereum/go-ethereum/crypto"
	lru "github.com/hashicorp/golang-lru/v2"
	"time"
)

// RequestAuthenticator authenticates requests to the DA node. This object is thread safe.
type RequestAuthenticator interface {
	// AuthenticateStoreChunksRequest authenticates a StoreChunksRequest, returning an error if the request is invalid.
	// The origin is the address of the peer that sent the request. This may be used to cache auth results
	// in order to save node resources.
	AuthenticateStoreChunksRequest(
		ctx context.Context,
		origin string,
		request *grpc.StoreChunksRequest,
		now time.Time) error
}

// keyWithTimeout is a key with that key's expiration time. After a key "expires", it should be reloaded
// from the chain state in case the key has been changed.
type keyWithTimeout struct {
	key        *ecdsa.PublicKey
	expiration time.Time
}

var _ RequestAuthenticator = &requestAuthenticator{}

type requestAuthenticator struct {
	// chainReader is used to read the chain state.
	chainReader core.Reader

	// keyCache is used to cache the public keys of dispersers.
	keyCache *lru.Cache[uint32, *keyWithTimeout]

	// keyTimeoutDuration is the duration for which a key is cached. After this duration, the key should be
	// reloaded from the chain state in case the key has been changed.
	keyTimeoutDuration time.Duration

	// authenticatedDispersers is a set of disperser addresses that have been recently authenticated, mapped
	// to the time when that cached authentication will expire.
	authenticatedDispersers *lru.Cache[string, time.Time]

	// authenticationTimeoutDuration is the duration for which an auth is valid.
	// If this is zero, then auth saving is disabled, and each request will be authenticated independently.
	authenticationTimeoutDuration time.Duration
}

// NewRequestAuthenticator creates a new RequestAuthenticator.
func NewRequestAuthenticator(
	ctx context.Context,
	chainReader core.Reader,
	keyCacheSize int,
	keyTimeoutDuration time.Duration,
	authenticationTimeoutDuration time.Duration,
	now time.Time) (RequestAuthenticator, error) {

	keyCache, err := lru.New[uint32, *keyWithTimeout](keyCacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create key cache: %w", err)
	}

	authenticatedDispersers, err := lru.New[string, time.Time](keyCacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticated dispersers cache: %w", err)
	}

	authenticator := &requestAuthenticator{
		chainReader:                   chainReader,
		keyCache:                      keyCache,
		keyTimeoutDuration:            keyTimeoutDuration,
		authenticatedDispersers:       authenticatedDispersers,
		authenticationTimeoutDuration: authenticationTimeoutDuration,
	}

	err = authenticator.preloadCache(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("failed to preload cache: %w", err)
	}

	return authenticator, nil
}

func (a *requestAuthenticator) preloadCache(ctx context.Context, now time.Time) error {
	// TODO (cody-littley): this will need to be updated for decentralized dispersers
	_, err := a.getDisperserKey(ctx, now, 0)
	if err != nil {
		return fmt.Errorf("failed to get operator key: %w", err)
	}

	return nil
}

func (a *requestAuthenticator) AuthenticateStoreChunksRequest(
	ctx context.Context,
	origin string,
	request *grpc.StoreChunksRequest,
	now time.Time) error {

	if a.isAuthenticationStillValid(now, origin) {
		// We've recently authenticated this client. Do not authenticate again for a while.
		return nil
	}

	key, err := a.getDisperserKey(ctx, now, request.DisperserID)
	if err != nil {
		return fmt.Errorf("failed to get operator key: %w", err)
	}

	signature := request.Signature
	isValid := dauth.VerifyStoreChunksRequest(key, request, signature)

	if !isValid {
		return errors.New("signature verification failed")
	}

	a.saveAuthenticationResult(now, origin)
	return nil
}

// getDisperserKey returns the public key of the operator with the given ID, caching the result.
func (a *requestAuthenticator) getDisperserKey(
	ctx context.Context,
	now time.Time,
	disperserID uint32) (*ecdsa.PublicKey, error) {
	key, ok := a.keyCache.Get(disperserID)
	if ok {
		expirationTime := key.expiration
		if now.Before(expirationTime) {
			return key.key, nil
		}
	}

	address, err := a.chainReader.GetDisperserAddress(ctx, disperserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get disperser address: %w", err)
	}

	ecdsaKey, err := crypto.UnmarshalPubkey(address.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	a.keyCache.Add(disperserID, &keyWithTimeout{
		key:        ecdsaKey,
		expiration: now.Add(a.keyTimeoutDuration),
	})

	return ecdsaKey, nil
}

// saveAuthenticationResult saves the result of an auth.
func (a *requestAuthenticator) saveAuthenticationResult(now time.Time, origin string) {
	if a.authenticationTimeoutDuration == 0 {
		// Authentication saving is disabled.
		return
	}

	a.authenticatedDispersers.Add(origin, now.Add(a.authenticationTimeoutDuration))
}

// isAuthenticationStillValid returns true if the client at the given address has been authenticated recently.
func (a *requestAuthenticator) isAuthenticationStillValid(now time.Time, address string) bool {
	if a.authenticationTimeoutDuration == 0 {
		// Authentication saving is disabled.
		return false
	}

	expiration, ok := a.authenticatedDispersers.Get(address)
	return ok && now.Before(expiration)
}
