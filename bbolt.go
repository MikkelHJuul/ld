package main

import (
	"io"

	bolt "go.etcd.io/bbolt"
)

type bboltDatabase struct {
	db           *bolt.DB
	bucketFinder func(*bolt.Tx) *bolt.Bucket //this is absolutely not as versatile as I want...
}

func InBucket(bucketFinder func(*bolt.Tx) *bolt.Bucket, f func(*bolt.Bucket) error) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		b := bucketFinder(tx)
		err := f(b)
		return err
	}
}

var _ kvDatabase = (*bboltDatabase)(nil)

// delete implements kvDatabase
func (b *bboltDatabase) delete(k key) (value, error) {
	var val value
	err := b.db.Update(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		val = b.Get(k)
		return b.Delete(k)
	}))
	if err != nil {
		return nil, err
	}
	return val, nil
}

// deleteMany implements kvDatabase
func (b *bboltDatabase) deleteMany(sen sender, rec receiver) error {
	return b.db.Update(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		for {
			keyV := rec.Receive()
			if keyV.error != nil && keyV.error == io.EOF {
				return nil
			} else if keyV.error != nil {
				return keyV.error
			}
			if err := sen.Send(kv{key: keyV.key, value: b.Get(keyV.key)}); err != nil {
				return err
			}
		}
	}))
}

// deleteRange implements kvDatabase
func (b *bboltDatabase) deleteRange(r ranger, sen sender) error {
	return b.db.Update(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		c := b.Cursor()
		for k, v := r.Start(c); k != nil && r.Valid(k); k, v = r.Next(c) {
			if r.Accept(k) {
				if err := c.Delete(); err != nil {
					return err
				}
				if err := sen.Send(kv{key: k, value: v}); err != nil {
					return err
				}
			}
		}
		return nil
	}))
}

// get implements kvDatabase
func (b *bboltDatabase) get(k key) (value, error) {
	var val value
	err := b.db.View(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		val = b.Get(k)
		return nil
	}))
	if err != nil {
		return nil, err
	}
	return val, nil
}

// getMany implements kvDatabase
func (b *bboltDatabase) getMany(sen sender, rec receiver) error {
	return b.db.View(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		for {
			keyV := rec.Receive()
			if keyV.error != nil && keyV.error == io.EOF {
				return nil
			} else if keyV.error != nil {
				return keyV.error
			}
			if err := sen.Send(kv{key: keyV.key, value: b.Get(keyV.key)}); err != nil {
				return err
			}
		}
	}))
}

// getRange implements kvDatabase
func (b *bboltDatabase) getRange(r ranger, sen sender) error {
	return b.db.View(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		c := b.Cursor()
		for k, v := r.Start(c); k != nil && r.Valid(k); k, v = r.Next(c) {
			if r.Accept(k) {
				if err := sen.Send(kv{key: k, value: v}); err != nil {
					return err
				}
			}
		}
		return nil
	}))
}

// set implements kvDatabase
func (b *bboltDatabase) set(keyV kv) (kv, error) {
	var val value
	err := b.db.Update(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		val = b.Get(keyV.key)
		return b.Put(keyV.key, keyV.value)
	}))
	if err != nil {
		return kv{}, err
	}
	return kv{key: keyV.key, value: val}, nil
}

// setMany implements kvDatabase
func (b *bboltDatabase) setMany(sen sender, rec receiver) error {
	return b.db.Update(InBucket(b.bucketFinder, func(b *bolt.Bucket) error {
		for {
			keyV := rec.Receive()
			if keyV.error != nil && keyV.error == io.EOF {
				return nil
			} else if keyV.error != nil {
				return keyV.error
			}
			old := kv{key: keyV.key, value: b.Get(keyV.key)}
			b.Put(keyV.key, keyV.value)
			if err := sen.Send(old); err != nil {
				return err
			}
		}
	}))
}
