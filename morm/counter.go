package morm

import (
	"github.com/iov-one/weave/orm"
)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (c *Counter) SetID(id []byte) error {
	c.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (c *Counter) Copy() orm.CloneableData {
	return &Counter{
		ID:    c.ID,
		Count: c.Count,
	}
}

// Validate is always succesful
func (c *Counter) Validate() error {
	return nil
}

var _ Model = (*Counter)(nil)
