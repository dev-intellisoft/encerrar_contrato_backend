package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/google/uuid"
	"github.com/tidwall/buntdb"
	"gorm.io/gorm"
)

// NewGormTokenStore keeps the original application contract from main.go.
// The project was missing the old storage package, so we provide a local
// token store implementation backed by buntdb for local execution.
func NewGormTokenStore(_ *gorm.DB) (oauth2.TokenStore, error) {
	return NewFileTokenStore("oauth_tokens.db")
}

func NewFileTokenStore(filename string) (oauth2.TokenStore, error) {
	db, err := buntdb.Open(filename)
	if err != nil {
		return nil, err
	}
	return &TokenStore{db: db}, nil
}

type TokenStore struct {
	db *buntdb.DB
}

func (ts *TokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	currentTime := time.Now()
	raw, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return ts.db.Update(func(tx *buntdb.Tx) error {
		if code := info.GetCode(); code != "" {
			_, _, err := tx.Set(
				code,
				string(raw),
				&buntdb.SetOptions{Expires: true, TTL: info.GetCodeExpiresIn()},
			)
			return err
		}

		basicID := uuid.Must(uuid.NewRandom()).String()
		accessTTL := info.GetAccessExpiresIn()
		refreshTTL := accessTTL
		expires := true

		if refresh := info.GetRefresh(); refresh != "" {
			refreshTTL = info.GetRefreshCreateAt().
				Add(info.GetRefreshExpiresIn()).
				Sub(currentTime)

			if accessTTL.Seconds() > refreshTTL.Seconds() {
				accessTTL = refreshTTL
			}

			expires = info.GetRefreshExpiresIn() != 0
			_, _, err := tx.Set(
				refresh,
				basicID,
				&buntdb.SetOptions{Expires: expires, TTL: refreshTTL},
			)
			if err != nil {
				return err
			}
		}

		_, _, err := tx.Set(
			basicID,
			string(raw),
			&buntdb.SetOptions{Expires: expires, TTL: refreshTTL},
		)
		if err != nil {
			return err
		}

		_, _, err = tx.Set(
			info.GetAccess(),
			basicID,
			&buntdb.SetOptions{Expires: expires, TTL: accessTTL},
		)
		return err
	})
}

func (ts *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	return ts.remove(code)
}

func (ts *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	return ts.remove(access)
}

func (ts *TokenStore) RemoveByRefresh(
	ctx context.Context,
	refresh string,
) error {
	return ts.remove(refresh)
}

func (ts *TokenStore) GetByCode(
	ctx context.Context,
	code string,
) (oauth2.TokenInfo, error) {
	return ts.getData(code)
}

func (ts *TokenStore) GetByAccess(
	ctx context.Context,
	access string,
) (oauth2.TokenInfo, error) {
	basicID, err := ts.getBasicID(access)
	if err != nil {
		return nil, err
	}
	return ts.getData(basicID)
}

func (ts *TokenStore) GetByRefresh(
	ctx context.Context,
	refresh string,
) (oauth2.TokenInfo, error) {
	basicID, err := ts.getBasicID(refresh)
	if err != nil {
		return nil, err
	}
	return ts.getData(basicID)
}

func (ts *TokenStore) remove(key string) error {
	err := ts.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		return err
	})
	if err == buntdb.ErrNotFound {
		return nil
	}
	return err
}

func (ts *TokenStore) getData(key string) (oauth2.TokenInfo, error) {
	var tokenInfo oauth2.TokenInfo
	err := ts.db.View(func(tx *buntdb.Tx) error {
		raw, err := tx.Get(key)
		if err != nil {
			return err
		}

		var token models.Token
		if err := json.Unmarshal([]byte(raw), &token); err != nil {
			return err
		}

		tokenInfo = &token
		return nil
	})
	if err != nil {
		if err == buntdb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return tokenInfo, nil
}

func (ts *TokenStore) getBasicID(key string) (string, error) {
	var basicID string
	err := ts.db.View(func(tx *buntdb.Tx) error {
		value, err := tx.Get(key)
		if err != nil {
			return err
		}
		basicID = value
		return nil
	})
	if err != nil {
		if err == buntdb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return basicID, nil
}
