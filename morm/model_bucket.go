package morm

import (
	"reflect"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

// TODO
// - migrations
// - do not use Bucket but directly access KVStore
// - register for queries

// Model is implemented by any entity that can be stored using ModelBucket.
//
// This is the same interface as CloneableData. Using the right type names
// provides an easier to read API.
//
// Model stores both the Key and the Value.
// GetID/SetID are used to store and access the Key.
// The ID is always set to nil before serializing and storing the Value.
type Model interface {
	weave.Persistent
	Validate() error
	Copy() orm.CloneableData
	GetID() []byte
	SetID([]byte) error
}

// ModelSlicePtr represents a pointer to a slice of models. Think of it as
// *[]Model Because of Go type system, using []Model would not work for us.
// Instead we use a placeholder type and the validation is done during the
// runtime.
type ModelSlicePtr interface{}

// ModelBucket is implemented by buckets that operates on Models rather than
// Objects.
type ModelBucket interface {
	// One query the database for a single model instance. Lookup is done
	// by the primary index key. Result is loaded into given destination
	// model.
	// This method returns ErrNotFound if the entity does not exist in the
	// database.
	// If given model type cannot be used to contain stored entity, ErrType
	// is returned.
	One(db weave.ReadOnlyKVStore, key []byte, dest Model) error

	// PrefixScan will scan for all models with a primary key (ID)
	// that begins with the given prefix.
	// The function returns a (possibly empty) iterator, which can
	// load each model as it arrives.
	// If reverse is true, iterates in descending order (highest value first),
	// otherwise, it iterates in
	PrefixScan(db weave.ReadOnlyKVStore, prefix []byte, reverse bool) (ModelIterator, error)

	// ByIndex returns all objects that secondary index with given name and
	// given key. Main index is always unique but secondary indexes can
	// return more than one value for the same key.
	// All matching entities are appended to given destination slice. If no
	// result was found, no error is returned and destination slice is not
	// modified.
	ByIndex(db weave.ReadOnlyKVStore, indexName string, key []byte, dest ModelSlicePtr) error

	// Put saves given model in the database. Before inserting into
	// database, model is validated using its Validate method.
	// If the key is nil or zero length then a sequence generator is used
	// to create a unique key value.
	// Using a key that already exists in the database cause the value to
	// be overwritten.
	Put(db weave.KVStore, m Model) error

	// Delete removes an entity with given primary key from the database.
	// It returns ErrNotFound if an entity with given key does not exist.
	Delete(db weave.KVStore, key []byte) error

	// Has returns nil if an entity with given primary key value exists. It
	// returns ErrNotFound if no entity can be found.
	// Has is a cheap operation that that does not read the data and only
	// checks the existence of it.
	Has(db weave.KVStore, key []byte) error

	// Register registers this buckets content to be accessible via query
	// requests under the given name.
	Register(name string, r weave.QueryRouter)
}

// NewModelBucket returns a ModelBucket instance. This implementation relies on
// a bucket instance. Final implementation should operate directly on the
// KVStore instead.
func NewModelBucket(name string, m Model, opts ...ModelBucketOption) ModelBucket {
	b := orm.NewBucket(name, orm.NewSimpleObj(nil, m))

	tp := reflect.TypeOf(m)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}

	mb := &modelBucket{
		b:     b,
		idSeq: b.Sequence("id"),
		model: tp,
	}
	for _, fn := range opts {
		fn(mb)
	}
	return mb
}

// ModelBucketOption is implemented by any function that can configure
// ModelBucket during creation.
type ModelBucketOption func(mb *modelBucket)

// WithIndex configures the bucket to build an index with given name. All
// entities stored in the bucket are indexed using value returned by the
// indexer function. If an index is unique, there can be only one entity
// referenced per index value.
func WithIndex(name string, indexer orm.Indexer, unique bool) ModelBucketOption {
	return func(mb *modelBucket) {
		mb.b = mb.b.WithIndex(name, indexer, unique)
	}
}

type modelBucket struct {
	b     orm.Bucket
	idSeq orm.Sequence

	// model is referencing the structure type. Event if the structure
	// pointer is implementing Model interface, this variable references
	// the structure directly and not the structure's pointer type.
	model reflect.Type
}

func (mb *modelBucket) Register(name string, r weave.QueryRouter) {
	mb.b.Register(name, r)
}

func (mb *modelBucket) One(db weave.ReadOnlyKVStore, key []byte, dest Model) error {
	obj, err := mb.b.Get(db, key)
	if err != nil {
		return err
	}
	if obj == nil || obj.Value() == nil {
		return errors.Wrapf(errors.ErrNotFound, "%T not in the store", dest)
	}
	res := obj.Value()

	if !reflect.TypeOf(res).AssignableTo(reflect.TypeOf(dest)) {
		return errors.Wrapf(errors.ErrType, "%T cannot be represented as %T", res, dest)
	}

	ptr := reflect.ValueOf(dest)
	ptr.Elem().Set(reflect.ValueOf(res).Elem())
	ptr.Interface().(Model).SetID(key)
	return nil
}

func (mb *modelBucket) PrefixScan(db weave.ReadOnlyKVStore, prefix []byte, reverse bool) (ModelIterator, error) {
	var rawIter weave.Iterator
	var err error

	start, end := prefixRange(mb.b.DBKey(prefix))
	if reverse {
		rawIter, err = db.ReverseIterator(start, end)
		if err != nil {
			return nil, errors.Wrap(err, "reverse prefix scan")
		}
	} else {
		rawIter, err = db.Iterator(start, end)
		if err != nil {
			return nil, errors.Wrap(err, "prefix scan")
		}
	}

	return &idModelIterator{iterator: rawIter, bucketPrefix: mb.b.DBKey(nil)}, nil
}

func (mb *modelBucket) ByIndex(db weave.ReadOnlyKVStore, indexName string, key []byte, destination ModelSlicePtr) error {
	objs, err := mb.b.GetIndexed(db, indexName, key)
	if err != nil {
		return err
	}
	if len(objs) == 0 {
		return nil
	}

	dest := reflect.ValueOf(destination)
	if dest.Kind() != reflect.Ptr {
		return errors.Wrap(errors.ErrType, "destination must be a pointer to slice of models")
	}
	if dest.IsNil() {
		return errors.Wrap(errors.ErrImmutable, "got nil pointer")
	}
	dest = dest.Elem()
	if dest.Kind() != reflect.Slice {
		return errors.Wrap(errors.ErrType, "destination must be a pointer to slice of models")
	}

	// It is allowed to pass destination as both []MyModel and []*MyModel
	sliceOfPointers := dest.Type().Elem().Kind() == reflect.Ptr

	allowed := dest.Type().Elem()
	if sliceOfPointers {
		allowed = allowed.Elem()
	}
	if mb.model != allowed {
		return errors.Wrapf(errors.ErrType, "this bucket operates on %s model and cannot return %s", mb.model, allowed)
	}

	for _, obj := range objs {
		if obj == nil || obj.Value() == nil {
			continue
		}
		val := reflect.ValueOf(obj.Value())
		val.Interface().(Model).SetID(obj.Key())
		if !sliceOfPointers {
			val = val.Elem()
		}
		// store the key on the model
		dest.Set(reflect.Append(dest, val))
	}
	return nil

}

func (mb *modelBucket) Put(db weave.KVStore, m Model) error {
	mTp := reflect.TypeOf(m)
	if mTp.Kind() != reflect.Ptr {
		return errors.Wrap(errors.ErrType, "model destination must be a pointer")
	}
	if mb.model != mTp.Elem() {
		return errors.Wrapf(errors.ErrType, "cannot store %T type in this bucket", m)
	}

	if err := m.Validate(); err != nil {
		return errors.Wrap(err, "invalid model")
	}

	key := m.GetID()
	if len(key) == 0 {
		var err error
		key, err = mb.idSeq.NextVal(db)
		if err != nil {
			return errors.Wrap(err, "ID sequence")
		}
	} else {
		// always nil out the key before saving the value
		m.SetID(nil)
	}

	obj := orm.NewSimpleObj(key, m)
	if err := mb.b.Save(db, obj); err != nil {
		return errors.Wrap(err, "cannot store in the database")
	}
	// after serialization, return original/generated key on model
	m.SetID(key)

	return nil
}

func (mb *modelBucket) Delete(db weave.KVStore, key []byte) error {
	if err := mb.Has(db, key); err != nil {
		return err
	}
	return mb.b.Delete(db, key)
}

func (mb *modelBucket) Has(db weave.KVStore, key []byte) error {
	if key == nil {
		// nil key is a special case that would cause the store API to panic.
		return errors.ErrNotFound
	}

	// As long as we rely on the Bucket implementation to access the
	// database, we must refine the key.
	key = mb.b.DBKey(key)

	ok, err := db.Has(key)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNotFound
	}
	return nil
}

var _ ModelBucket = (*modelBucket)(nil)
