package filecache

import (
	"errors"
	"github.com/goccy/go-json"
	bolt "go.etcd.io/bbolt"
	"time"
)

// Reference: https://github.com/metafates/mangal/blob/v5/cache/bbolt/bbolt.go

const ttlBucketName = "ttl"

type Store struct {
	db            *bolt.DB
	ttl           time.Duration
	ttlBucketName string
	bucketName    string
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (s Store) Set(k string, v interface{}) error {
	if err := checkKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that bbolt can handle
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		if err := b.Put([]byte(k), data); err != nil {
			return err
		}

		bTTL := tx.Bucket([]byte(s.ttlBucketName))
		return bTTL.Put([]byte(k), []byte(time.Now().UTC().Format(time.RFC3339Nano)))
	})
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s Store) Get(k string, v interface{}) (found bool, err error) {
	if err := checkKeyAndValue(k, v); err != nil {
		return false, err
	}

	var data []byte
	err = s.db.View(func(tx *bolt.Tx) error {
		bTTL := tx.Bucket([]byte(s.ttlBucketName))
		// Get the TTL for the key
		ttlData := bTTL.Get([]byte(k))
		if ttlData != nil {
			ttl, err := time.Parse(time.RFC3339Nano, string(ttlData))
			if err != nil {
				return err
			}
			if time.Now().UTC().After(ttl.Add(s.ttl)) {
				return nil
			}
		}

		b := tx.Bucket([]byte(s.bucketName))
		txData := b.Get([]byte(k))
		// txData is only valid during the transaction.
		// Its value must be copied to make it valid outside of the tx.
		// TODO: Benchmark if it's faster to copy + close tx,
		// or to keep the tx open until unmarshalling is done.
		if txData != nil {
			// `data = append([]byte{}, txData...)` would also work, but the following is more explicit
			data = make([]byte, len(txData))
			copy(data, txData)
		}
		return nil
	})
	if err != nil {
		return false, nil
	}

	// If no value was found return false
	if data == nil {
		return false, nil
	}

	return true, json.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (s Store) Delete(k string) error {
	if err := checkKey(k); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.bucketName))
		if err := b.Delete([]byte(k)); err != nil {
			return err
		}

		bTTL := tx.Bucket([]byte(s.ttlBucketName))
		return bTTL.Delete([]byte(k))
	})
}

// Close closes the store.
// It must be called to make sure that all open transactions finish and to release all DB resources.
func (s Store) Close() error {
	return s.db.Close()
}

// Options are the options for the bbolt store.
type Options struct {
	TTL time.Duration
	// Bucket name for storing the key-value pairs.
	// Optional ("default" by default).
	BucketName string
	// Path of the DB file.
	// Optional ("bbolt.db" by default).
	Path string
	// Encoding format.
	// Optional (encoding.JSON by default).
}

// DefaultOptions is an Options object with default values.
var DefaultOptions = Options{
	TTL:        time.Hour * 24 * 7,
	BucketName: "default",
	Path:       "cache.db",
}

// NewStore creates a new bbolt store.
// Note: bbolt uses an exclusive write lock on the database file, so it cannot be shared by multiple processes.
// So when creating multiple clients you should always use a new database file (by setting a different Path in the options).
//
// You must call the Close() method on the store when you're done working with it.
func NewStore(options Options) (*Store, error) {
	result := Store{}

	// Set default values
	if options.BucketName == "" {
		options.BucketName = DefaultOptions.BucketName
	}
	if options.Path == "" {
		options.Path = DefaultOptions.Path
	}
	if options.TTL == 0 {
		options.TTL = DefaultOptions.TTL
	}

	// Open DB
	db, err := bolt.Open(options.Path, 0600, nil)
	if err != nil {
		return &result, err
	}

	// Create a bucket if it doesn't exist yet.
	// In bbolt key/value pairs are stored to and read from buckets.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(options.BucketName))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(ttlBucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return &result, err
	}

	result.ttl = options.TTL
	result.db = db
	result.ttlBucketName = ttlBucketName
	result.bucketName = options.BucketName

	return &result, nil
}

// checkKeyAndValue returns an error if k == "" or if v == nil
func checkKeyAndValue(k string, v any) error {
	if err := checkKey(k); err != nil {
		return err
	}
	return checkVal(v)
}

// checkKey returns an error if k == ""
func checkKey(k string) error {
	if k == "" {
		return errors.New("the passed key is an empty string, which is invalid")
	}
	return nil
}

// checkVal returns an error if v == nil
func checkVal(v any) error {
	if v == nil {
		return errors.New("the passed value is nil, which is not allowed")
	}
	return nil
}
