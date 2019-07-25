package morm

import (
	"bytes"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

type ModelIterator interface {
	// Next moves the iterator to the next sequential key in the database, as
	// defined by order of iteration.
	Next() (key, value []byte, err error)

	// Load reads the current value at the given key into the passed destination.
	// This works much like "One" in ModelBucket
	Load(key, value []byte, dest Model) error

	// Release releases the Iterator.
	Release()
}

type idModelIterator struct {
	// this is the raw KVStoreIterator
	iterator weave.Iterator
	// this is the bucketPrefix to strip from each key
	bucketPrefix []byte
}

var _ ModelIterator = (*idModelIterator)(nil)

func (i *idModelIterator) Next() (key, value []byte, err error) {
	return i.iterator.Next()
}

func (i *idModelIterator) Release() {
	i.iterator.Release()
}

func (i *idModelIterator) Load(key, value []byte, dest Model) error {
	return load(key, value, i.bucketPrefix, dest)
}

type indexModelIterator struct {
	// this is the raw KVStoreIterator
	iterator weave.Iterator
	// this is the bucketPrefix to strip from each key
	bucketPrefix []byte
	unique       bool

	kv         weave.ReadOnlyKVStore
	cachedKeys [][]byte
}

var _ ModelIterator = (*indexModelIterator)(nil)

func (i *indexModelIterator) Next() (key, value []byte, err error) {
	if len(i.cachedKeys) > 1 {
		i.cachedKeys = i.cachedKeys[1:]
	} else {
		i.cachedKeys = nil
	}

	return i.iterator.Next()
}

func (i *indexModelIterator) Release() {
	i.iterator.Release()
}

func (i *indexModelIterator) Load(key, value []byte,dest Model) error {
	// if we have cached keys, just use those, not the iterator value
	if len(i.cachedKeys) > 0 {
		key = i.dbKey(i.cachedKeys[0])
	} else {
		keys, err := i.getRefs(value, i.unique)
		if err != nil {
			return errors.Wrap(err, "parsing index refs")
		}

		if len(keys) != 1 {
			i.cachedKeys = keys
		}
		key = i.dbKey(keys[0])
	}

	val, err := i.kv.Get(key)
	if err != nil {
		return errors.Wrap(err, "loading referenced key")
	}
	if val == nil {
		return errors.Wrapf(errors.ErrNotFound, "key: %X", key)
	}
	return load(key, val, i.bucketPrefix, dest)
}

// get refs takes a value stored in an index and parse it into a slice of
// db keys
func (i *indexModelIterator) getRefs(val []byte, unique bool) ([][]byte, error) {
	if val == nil {
		return nil, nil
	}
	if unique {
		return [][]byte{val}, nil
	}
	var data = new(orm.MultiRef)
	err := data.Unmarshal(val)
	if err != nil {
		return nil, err
	}
	return data.GetRefs(), nil
}

func (i *indexModelIterator) dbKey(key []byte) []byte {
	return append(i.bucketPrefix, key...)
}



func load(key, value, bucketPrefix []byte, dest Model) error {
	// since we use raw kvstore here, not Bucket, we must remove the bucket prefix manually
	if !bytes.HasPrefix(key, bucketPrefix) {
		return errors.Wrapf(errors.ErrDatabase, "key with unexpected prefix: %X", key)
	}
	key = key[len(bucketPrefix):]

	if err := dest.Unmarshal(value); err != nil {
		return errors.Wrapf(err, "unmarshaling into %T", dest)
	}
	if err := dest.SetID(key); err != nil {
		return errors.Wrap(err, "setting ID")
	}
	return nil
}

// prefixRange turns a prefix into (start, end) to create
// and iterator
func prefixRange(prefix []byte) ([]byte, []byte) {
	// special case: no prefix is whole range
	if len(prefix) == 0 {
		return nil, nil
	}

	// copy the prefix and update last byte
	end := make([]byte, len(prefix))
	copy(end, prefix)
	l := len(end) - 1
	end[l]++

	// wait, what if that overflowed?....
	for end[l] == 0 && l > 0 {
		l--
		end[l]++
	}

	// okay, funny guy, you gave us FFF, no end to this range...
	if l == 0 && end[0] == 0 {
		end = nil
	}
	return prefix, end
}
