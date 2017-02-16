package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/atvaark/dragons-dogma-server/modules/game"
	"github.com/boltdb/bolt"
)

type Database interface {
	io.Closer
	GetOnlineUrDragon() (*game.OnlineUrDragon, error)
	PutOnlineUrDragon(*game.OnlineUrDragon) error
}

type boltDB struct {
	innerDB *bolt.DB
}

func NewDatabase(path string) (Database, error) {
	boltOptions := &bolt.Options{Timeout: 10 * time.Second}
	innerDB, err := bolt.Open(path, 0600, boltOptions)
	if err != nil {
		return nil, err
	}

	database := &boltDB{innerDB: innerDB}
	err = database.init()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (db *boltDB) init() error {
	err := db.innerDB.Update(func(tx *bolt.Tx) error {
		err := initDragonBucket(tx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not initialize the database: %v", err)
	}

	return nil
}

func initDragonBucket(tx *bolt.Tx) error {
	b := tx.Bucket(dragonBucketName)
	if b == nil {
		b, err := tx.CreateBucket(dragonBucketName)
		if err != nil {
			return err
		}

		d := game.NewOnlineUrDragon()
		err = putOnlineUrDragonInternal(b, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *boltDB) Close() error {
	err := db.innerDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close the database: %v", err)
	}

	return nil
}

var (
	dragonBucketName = []byte("dragon")
	dragonBucketKey  = []byte("dragon")
)

func (db *boltDB) GetOnlineUrDragon() (dragon *game.OnlineUrDragon, err error) {
	err = db.innerDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(dragonBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		v := b.Get(dragonBucketKey)
		if v == nil {
			return errors.New("dragon not found")
		}

		var d game.OnlineUrDragon
		err = json.Unmarshal(v, &d)
		if err != nil {
			return err
		}

		dragon = &d

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not retrieve the online ur dragon: %v", err)
	}

	return dragon, nil
}

func (db *boltDB) PutOnlineUrDragon(dragon *game.OnlineUrDragon) error {
	err := db.innerDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(dragonBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		err := putOnlineUrDragonInternal(b, dragon)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not save the online ur dragon: %v", err)
	}

	return nil
}

func putOnlineUrDragonInternal(b *bolt.Bucket, dragon *game.OnlineUrDragon) error {
	v, err := json.Marshal(dragon)
	if err != nil {
		return err
	}

	err = b.Put(dragonBucketKey, v)
	if err != nil {
		return err
	}

	return nil
}
