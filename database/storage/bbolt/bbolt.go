package bbolt

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"

	"github.com/safing/portbase/database/iterator"
	"github.com/safing/portbase/database/query"
	"github.com/safing/portbase/database/record"
	"github.com/safing/portbase/database/storage"
)

var (
	bucketName = []byte{0}
)

// BBolt database made pluggable for portbase.
type BBolt struct {
	name string
	db   *bbolt.DB
}

func init() {
	_ = storage.Register("bbolt", NewBBolt)
}

// NewBBolt opens/creates a bbolt database.
func NewBBolt(name, location string) (storage.Interface, error) {

	db, err := bbolt.Open(filepath.Join(location, "db.bbolt"), 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &BBolt{
		name: name,
		db:   db,
	}, nil
}

// Get returns a database record.
func (b *BBolt) Get(key string) (record.Record, error) {
	var r record.Record

	err := b.db.View(func(tx *bbolt.Tx) error {
		// get value from db
		value := tx.Bucket(bucketName).Get([]byte(key))
		if value == nil {
			return storage.ErrNotFound
		}

		// copy data
		duplicate := make([]byte, len(value))
		copy(duplicate, value)

		// create record
		var txErr error
		r, txErr = record.NewRawWrapper(b.name, key, duplicate)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return r, nil
}

// Put stores a record in the database.
func (b *BBolt) Put(r record.Record) error {
	data, err := r.MarshalRecord(r)
	if err != nil {
		return err
	}

	err = b.db.Update(func(tx *bbolt.Tx) error {
		txErr := tx.Bucket(bucketName).Put([]byte(r.DatabaseKey()), data)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a record from the database.
func (b *BBolt) Delete(key string) error {
	err := b.db.Update(func(tx *bbolt.Tx) error {
		txErr := tx.Bucket(bucketName).Delete([]byte(key))
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Query returns a an iterator for the supplied query.
func (b *BBolt) Query(q *query.Query, local, internal bool) (*iterator.Iterator, error) {
	_, err := q.Check()
	if err != nil {
		return nil, fmt.Errorf("invalid query: %s", err)
	}

	queryIter := iterator.New()

	go b.queryExecutor(queryIter, q, local, internal)
	return queryIter, nil
}

func (b *BBolt) queryExecutor(queryIter *iterator.Iterator, q *query.Query, local, internal bool) {
	prefix := []byte(q.DatabaseKeyPrefix())
	err := b.db.View(func(tx *bbolt.Tx) error {
		// Create a cursor for iteration.
		c := tx.Bucket(bucketName).Cursor()

		// Iterate over items in sorted key order. This starts from the
		// first key/value pair and updates the k/v variables to the
		// next key/value on each iteration.
		//
		// The loop finishes at the end of the cursor when a nil key is returned.
		for key, value := c.Seek(prefix); key != nil; key, value = c.Next() {

			// if we don't match the prefix anymore, exit
			if !bytes.HasPrefix(key, prefix) {
				return nil
			}

			// wrap value
			iterWrapper, err := record.NewRawWrapper(b.name, string(key), value)
			if err != nil {
				return err
			}

			// check validity / access
			if !iterWrapper.Meta().CheckValidity() {
				continue
			}
			if !iterWrapper.Meta().CheckPermission(local, internal) {
				continue
			}

			// check if matches & send
			if q.MatchesRecord(iterWrapper) {
				// copy data
				duplicate := make([]byte, len(value))
				copy(duplicate, value)

				new, err := record.NewRawWrapper(b.name, iterWrapper.DatabaseKey(), duplicate)
				if err != nil {
					return err
				}
				select {
				case <-queryIter.Done:
					return nil
				case queryIter.Next <- new:
				default:
					select {
					case <-queryIter.Done:
						return nil
					case queryIter.Next <- new:
					case <-time.After(1 * time.Second):
						return errors.New("query timeout")
					}
				}
			}
		}
		return nil
	})
	queryIter.Finish(err)
}

// ReadOnly returns whether the database is read only.
func (b *BBolt) ReadOnly() bool {
	return false
}

// Injected returns whether the database is injected.
func (b *BBolt) Injected() bool {
	return false
}

// Maintain runs a light maintenance operation on the database.
func (b *BBolt) Maintain() error {
	return nil
}

// MaintainThorough runs a thorough maintenance operation on the database.
func (b *BBolt) MaintainThorough() (err error) {
	return nil
}

// Shutdown shuts down the database.
func (b *BBolt) Shutdown() error {
	return b.db.Close()
}
