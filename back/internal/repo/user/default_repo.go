package user

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/patrickmn/go-cache"
)

type DefaultRepo struct {
	db    *bolt.DB
	cache *cache.Cache
}

var (
	UserBucket = []byte("user")
)

type User struct {
	Email        string    `json:"email,omitempty"`
	PassHash     string    `json:"pass_hash,omitempty"`
	Salt         string    `json:"salt"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	ValidEmail   bool      `json:"valid_email,omitempty"`
	ActiveJWT    []string  `json:"active_jwt,omitempty"`
	Notes        []string  `json:"notes,omitempty"`
	SharedWithMe []string  `json:"shared_with_me,omitempty"`
}

func NewDefaultRepo(db *bolt.DB, cache *cache.Cache) (*DefaultRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(UserBucket); err != nil {
			return err
		}
		return nil
	})
	return &DefaultRepo{
		db:    db,
		cache: cache,
	}, err
}

func (r *DefaultRepo) GetUser(email string) (*User, error) {
	var user User
	cachedUser, found := r.cache.Get(email)
	if found {
		return cachedUser.(*User), nil
	}

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		if b == nil {
			return fmt.Errorf("user bucket not found")
		}

		userBytes := b.Get([]byte(email))
		if userBytes == nil {
			return fmt.Errorf("user not found")
		}

		err := json.Unmarshal(userBytes, &user)
		if err != nil {
			return fmt.Errorf("failed to unmarshal user > %w", err)
		}

		return nil
	})

	r.cache.Set(email, &user, cache.DefaultExpiration)
	return &user, err
}

func (r *DefaultRepo) SaveUser(u *User) error {
	userBytes, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("failed to marshal user > %w", err)
	}

	r.cache.Delete(u.Email)

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		if b == nil {
			return fmt.Errorf("user bucket not found")
		}
		return b.Put([]byte(u.Email), userBytes)
	})
	if err != nil {
		return fmt.Errorf("failed to store user in db > %w", err)
	}

	return nil
}

func (r *DefaultRepo) DeleteUser(email string) error {
	r.cache.Delete(email)

	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		if b == nil {
			return fmt.Errorf("user bucket not found")
		}
		return b.Delete([]byte(email))
	})
	if err != nil {
		return fmt.Errorf("failed to delete user > %w", err)
	}

	return nil
}
