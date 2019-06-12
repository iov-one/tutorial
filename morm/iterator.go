package morm

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
)

type ModelIterator interface {
	// Valid returns whether the current position is valid.
	// Once invalid, an Iterator is forever invalid.
	Valid() bool

	// Next moves the iterator to the next sequential key in the database, as
	// defined by order of iteration.
	//
	// If Valid returns false, this method will panic.
	Next() error

	// Load reads the current value at the given key into the passed destination.
	// This works much like "One" in ModelBucket
	Load(dest Model) error

	// Close releases the Iterator.
	Close()
}

type idModelIterator struct {
	iterator weave.Iterator
}

var _ ModelIterator = (*idModelIterator)(nil)

func (i *idModelIterator) Valid() bool {
	return i.iterator.Valid()
}

func (i *idModelIterator) Next() error {
	return i.iterator.Next()
}

func (i *idModelIterator) Close() {
	i.iterator.Close()
}

func (i *idModelIterator) Load(dest Model) error {
	key := i.iterator.Key()
	value := i.iterator.Value()

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
