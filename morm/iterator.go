package morm

import (
	"bytes"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

type ModelIterator interface {
	// LoadNext moves the iterator to the next sequntial key in the database and
	// loads the current value at the given key into the passed destination.
	LoadNext(dest Model) error

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

func (i *idModelIterator) LoadNext(dest Model) error {
	key, value, err := i.iterator.Next()
	if err != nil {
		return err
	}

	return load(key, value, i.bucketPrefix, dest)
}

func (i *idModelIterator) Release() {
	i.iterator.Release()
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

// LoadNext loads next iterator value to dest
func (i *indexModelIterator) LoadNext(dest Model) error {
	key, err := i.getKey()
	if err != nil {
		return err
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

func (i *indexModelIterator) Release() {
	i.iterator.Release()
}

// getKey retrieves the key from cache if i.cacheKeys is not nil, otherwise loads next iterator key
func (i *indexModelIterator) getKey() ([]byte, error) {
	var key []byte

	switch cachedKeysLen := len(i.cachedKeys); {
	case cachedKeysLen > 1:
		//gets the key from cache and remove first key from i.cacheKey
		key = i.dbKey(i.cachedKeys[0])
		i.cachedKeys = i.cachedKeys[1:]
	case cachedKeysLen == 1:
		//gets the key from cache and sets i.cachedKeys as nil
		key = i.dbKey(i.cachedKeys[0])
		i.cachedKeys = nil
	default:
		//retrievesthe key and value from iterator
		_, value, err := i.iterator.Next()
		if err != nil {
			return nil, err
		}

		keys, err := i.getRefs(value, i.unique)
		if err != nil {
			return nil, errors.Wrap(err, "parsing index refs")
		}
		if len(keys) != 1 {
			i.cachedKeys = keys[1:]
		}
		key = i.dbKey(keys[0])
	}

	return key, nil
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
