package auth

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

type DefaultRepo struct {
	db *bolt.DB
}

var (
	UserBucket = []byte("user")
)

func NewDefaultRepo(db *bolt.DB) (*DefaultRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(UserBucket)
		return err
	})
	return &DefaultRepo{
		db: db,
	}, err
}

func (r *DefaultRepo) SaveUser(user *UserClaim) error {
	userBytes, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user > %w", err)
	}

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		return b.Put([]byte(user.Email), userBytes)
	})
	if err != nil {
		return fmt.Errorf("failed to store user in db > %w", err)
	}

	return nil
}

func (r *DefaultRepo) GetUser(email string) (*UserClaim, error) {
	var user UserClaim
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		userBytes := b.Get([]byte(email))
		if len(userBytes) == 0 {
			return fmt.Errorf("user not found")
		}

		err := json.Unmarshal(userBytes, &user)
		if err != nil {
			return fmt.Errorf("failed to parse user > %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}
