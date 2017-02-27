package db

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/atvaark/dragons-dogma-server/modules/auth"
	"github.com/atvaark/dragons-dogma-server/modules/game"
	"github.com/boltdb/bolt"
)

type Database interface {
	io.Closer
	game.Database
	auth.Database
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

func (db *boltDB) Close() error {
	err := db.innerDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close the database: %v", err)
	}

	return nil
}

func (db *boltDB) init() error {
	err := db.innerDB.Update(func(tx *bolt.Tx) error {
		var err error
		err = initDragonBucket(tx)
		if err != nil {
			return err
		}

		err = initPawnRewardBucket(tx)
		if err != nil {
			return err
		}

		err = initSessionBucket(tx)
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

var (
	dragonBucketName = []byte("dragon")
	dragonBucketKey  = []byte("dragon")
)

func initDragonBucket(tx *bolt.Tx) error {
	b := tx.Bucket(dragonBucketName)
	if b == nil {
		b, err := tx.CreateBucket(dragonBucketName)
		if err != nil {
			return err
		}

		d := &game.OnlineUrDragon{}
		d = d.NextGeneration()
		err = putOnlineUrDragonInternal(b, d)
		if err != nil {
			return err
		}
	}

	return nil
}

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

var (
	sessionBucketName = []byte("session")
)

func initSessionBucket(tx *bolt.Tx) error {
	b := tx.Bucket(sessionBucketName)
	if b == nil {
		_, err := tx.CreateBucket(sessionBucketName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *boltDB) GetSession(ID string) (session *auth.Session, err error) {
	err = db.innerDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		v := b.Get([]byte(ID))
		if v == nil {
			return fmt.Errorf("session '%s' not found", ID)
		}

		var s auth.Session
		err = json.Unmarshal(v, &s)
		if err != nil {
			return err
		}

		session = &s

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not retrieve the session: %v", err)
	}

	return session, nil
}

func (db *boltDB) PutSession(session *auth.Session) error {
	err := db.innerDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		v, err := json.Marshal(session)
		if err != nil {
			return err
		}

		err = b.Put([]byte(session.ID), v)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not save the session: %v", err)
	}

	return nil
}

func (db *boltDB) DeleteSession(ID string) error {
	err := db.innerDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		err := b.Delete([]byte(ID))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not delete the session: %v", err)
	}

	return nil
}

var (
	pawnRewardBucketName = []byte("pawnreward")
)

func initPawnRewardBucket(tx *bolt.Tx) error {
	b := tx.Bucket(pawnRewardBucketName)
	if b == nil {
		_, err := tx.CreateBucket(pawnRewardBucketName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *boltDB) GetPawnRewards(userID uint64) (rewards *game.PawnRewards, err error) {
	err = db.innerDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(pawnRewardBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		v := b.Get(userIDToKey(userID))
		if v == nil {
			return nil
		}

		var r game.PawnRewards
		err = json.Unmarshal(v, &r)
		if err != nil {
			return err
		}

		rewards = &r

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not retrieve the pawn rewards: %v", err)
	}

	return rewards, nil
}

func (db *boltDB) PutPawnRewards(rewards *game.PawnRewards) error {
	err := db.innerDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(pawnRewardBucketName)
		if b == nil {
			return errors.New("database not initialized")
		}

		v, err := json.Marshal(rewards)
		if err != nil {
			return err
		}

		err = b.Put(userIDToKey(rewards.PawnUserID), v)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not save the pawn rewards: %v", err)
	}

	return nil
}

func userIDToKey(userID uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b[:], userID)
	return b
}
