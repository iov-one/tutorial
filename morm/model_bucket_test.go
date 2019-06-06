package morm

import (
	"strconv"
	"testing"

	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"github.com/iov-one/weave/store"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

func TestModelBucket(t *testing.T) {
	db := store.MemStore()

	b := NewModelBucket("cnts", &Counter{})

	key1 := []byte("c1")
	err := b.Put(db, &Counter{ID: key1, Count: 1})
	assert.Nil(t, err)

	var c1 Counter
	err = b.One(db, key1, &c1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), c1.Count)
	assert.Equal(t, key1, c1.GetID())

	err = b.Delete(db, key1)
	assert.Nil(t, err)
	if err := b.Delete(db, []byte("unknown")); !errors.ErrNotFound.Is(err) {
		t.Fatalf("unexpected error when deleting unexisting instance: %s", err)
	}
	if err := b.One(db, key1, &c1); !errors.ErrNotFound.Is(err) {
		t.Fatalf("unexpected error for an unknown model get: %s", err)
	}
}

func TestModelBucketPutSequence(t *testing.T) {
	db := store.MemStore()

	b := NewModelBucket("cnts", &Counter{})

	// Using a nil key should cause the sequence ID to be used.
	cnt := Counter{Count: 111}
	assert.Nil(t, cnt.GetID())
	err := b.Put(db, &cnt)
	assert.Nil(t, err)
	assert.Equal(t, cnt.GetID(), weavetest.SequenceID(1))

	// Inserting an entity with a key provided must not modify the ID
	// generation counter.
	err = b.Put(db, &Counter{ID: []byte("mycnt"), Count: 12345})
	assert.Nil(t, err)

	cnt2 := Counter{Count: 222}
	err = b.Put(db, &cnt2)
	assert.Nil(t, err)
	assert.Equal(t, cnt2.GetID(), weavetest.SequenceID(2))

	var c1 Counter
	err = b.One(db, weavetest.SequenceID(1), &c1)
	assert.Nil(t, err)
	assert.Equal(t, weavetest.SequenceID(1), c1.GetID())
	assert.Equal(t, int64(111), c1.Count)

	var c2 Counter
	err = b.One(db, weavetest.SequenceID(2), &c2)
	assert.Nil(t, err)
	assert.Equal(t, weavetest.SequenceID(2), c2.GetID())
	assert.Equal(t, int64(222), c2.Count)
}

func TestModelBucketByIndex(t *testing.T) {
	cases := map[string]struct {
		IndexName  string
		QueryKey   string
		DestFn     func() ModelSlicePtr
		WantErr    *errors.Error
		WantResPtr []*Counter
		WantRes    []Counter
	}{
		"find none": {
			IndexName:  "value",
			QueryKey:   "124089710947120",
			WantErr:    nil,
			WantResPtr: nil,
			WantRes:    nil,
		},
		"find one": {
			IndexName: "value",
			QueryKey:  "1",
			WantErr:   nil,
			WantResPtr: []*Counter{
				{
					ID:    weavetest.SequenceID(1),
					Count: 1001,
				},
			},
			WantRes: []Counter{
				{
					ID:    weavetest.SequenceID(1),
					Count: 1001,
				},
			},
		},
		"find two": {
			IndexName: "value",
			QueryKey:  "4",
			WantErr:   nil,
			WantResPtr: []*Counter{
				{ID: weavetest.SequenceID(3), Count: 4001},
				{ID: weavetest.SequenceID(4), Count: 4002},
			},
			WantRes: []Counter{
				{ID: weavetest.SequenceID(3), Count: 4001},
				{ID: weavetest.SequenceID(4), Count: 4002},
			},
		},
		"non existing index name": {
			IndexName: "xyz",
			WantErr:   orm.ErrInvalidIndex,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			db := store.MemStore()

			indexByBigValue := func(obj orm.Object) ([]byte, error) {
				c, ok := obj.Value().(*Counter)
				if !ok {
					return nil, errors.Wrapf(errors.ErrType, "%T", obj.Value())
				}
				// Index by the value, ignoring anything below 1k.
				raw := strconv.FormatInt(c.Count/1000, 10)
				return []byte(raw), nil
			}
			b := NewModelBucket("cnts", &Counter{}, WithIndex("value", indexByBigValue, false))

			err := b.Put(db, &Counter{Count: 1001})
			assert.Nil(t, err)
			err = b.Put(db, &Counter{Count: 2001})
			assert.Nil(t, err)
			err = b.Put(db, &Counter{Count: 4001})
			assert.Nil(t, err)
			err = b.Put(db, &Counter{Count: 4002})
			assert.Nil(t, err)

			var dest []Counter
			err = b.ByIndex(db, tc.IndexName, []byte(tc.QueryKey), &dest)
			if !tc.WantErr.Is(err) {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.Equal(t, tc.WantRes, dest)

			var destPtr []*Counter
			err = b.ByIndex(db, tc.IndexName, []byte(tc.QueryKey), &destPtr)
			if !tc.WantErr.Is(err) {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.Equal(t, tc.WantResPtr, destPtr)
		})
	}
}

// BadCounter implements Model but is different type than the model
type BadCounter struct {
	Counter
	Random int
}

var _ Model = (*BadCounter)(nil)

func TestModelBucketPutWrongModelType(t *testing.T) {
	db := store.MemStore()
	b := NewModelBucket("cnts", &Counter{})

	if err := b.Put(db, &BadCounter{Counter: Counter{Count: 5}, Random: 77}); !errors.ErrType.Is(err) {
		t.Fatalf("unexpected error when trying to store wrong model type value: %s", err)
	}
}

func TestModelBucketOneWrongModelType(t *testing.T) {
	db := store.MemStore()
	b := NewModelBucket("cnts", &Counter{})

	err := b.Put(db, &Counter{ID: []byte("counter"), Count: 1})
	assert.Nil(t, err)

	var bad BadCounter
	if err := b.One(db, []byte("counter"), &bad); !errors.ErrType.Is(err) {
		t.Fatalf("unexpected error when trying to get wrong model type value: %s", err)
	}
}

func TestModelBucketByIndexWrongModelType(t *testing.T) {
	db := store.MemStore()
	indexer := func(o orm.Object) ([]byte, error) {
		return []byte("x"), nil
	}
	b := NewModelBucket("cnts", &Counter{}, WithIndex("x", indexer, false))

	err := b.Put(db, &Counter{ID: []byte("counter"), Count: 1})
	assert.Nil(t, err)

	var bads []BadCounter
	if err := b.ByIndex(db, "x", []byte("x"), &bads); !errors.ErrType.Is(err) {
		t.Fatalf("unexpected error when trying to find wrong model type value: %s: %v", err, bads)
	}

	var badsPtr []*BadCounter
	if err := b.ByIndex(db, "x", []byte("x"), &badsPtr); !errors.ErrType.Is(err) {
		t.Fatalf("unexpected error when trying to find wrong model type value: %s: %v", err, badsPtr)
	}

	var badsPtrPtr []**BadCounter
	if err := b.ByIndex(db, "x", []byte("x"), &badsPtrPtr); !errors.ErrType.Is(err) {
		t.Fatalf("unexpected error when trying to find wrong model type value: %s: %v", err, badsPtrPtr)
	}
}

func TestModelBucketHas(t *testing.T) {
	db := store.MemStore()
	b := NewModelBucket("cnts", &Counter{})

	err := b.Put(db, &Counter{ID: []byte("counter"), Count: 1})
	assert.Nil(t, err)

	err = b.Has(db, []byte("counter"))
	assert.Nil(t, err)

	if err := b.Has(db, nil); !errors.ErrNotFound.Is(err) {
		t.Fatalf("a nil key must return ErrNotFound: %s", err)
	}

	if err := b.Has(db, []byte("does-not-exist")); !errors.ErrNotFound.Is(err) {
		t.Fatalf("a non exists entity must return ErrNotFound: %s", err)
	}
}
