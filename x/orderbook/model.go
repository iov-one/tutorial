package orderbook

import (
	"regexp"

	"github.com/iov-one/tutorial/morm"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

var _ morm.Model = (*Market)(nil)

var validMarketName = regexp.MustCompile(`^[a-zA-Z0-9_.-]{4,32}$`).MatchString

func copyBytes(in []byte) []byte {
	if in == nil {
		return nil
	}
	cpy := make([]byte, len(in))
	copy(cpy, in)
	return cpy
}

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (m *Market) SetID(id []byte) error {
	m.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (m *Market) Copy() orm.CloneableData {
	return &Market{
		Metadata: m.Metadata.Copy(),
		ID:       copyBytes(m.ID),
		Owner:    m.Owner.Clone(),
		Name:     m.Name,
	}
}

// Validate is always succesful
func (m *Market) Validate() error {
	if err := m.Metadata.Validate(); err != nil {
		return err
	}
	if err := m.Owner.Validate(); err != nil {
		return errors.Wrap(err, "owner")
	}
	if !validMarketName(m.Name) {
		return errors.Wrap(errors.ErrModel, "invalid market name")
	}
	return nil
}
