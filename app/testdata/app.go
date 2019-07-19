package fixtures

import (
	"fmt"
	"math/rand"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/crypto"
)

type AppFixture struct {
	Name              string
	ChainID           string
	GenesisKey        *crypto.PrivateKey
	GenesisKeyAddress weave.Address
}

func NewApp() *AppFixture {
	pk := crypto.GenPrivKeyEd25519()
	addr := pk.PublicKey().Address()
	name := fmt.Sprint("test-%d", rand.Intn(99999999))
	return &AppFixture{
		Name:              name,
		ChainID:           fmt.Sprint("chain-%s", name),
		GenesisKey:        pk,
		GenesisKeyAddress: addr,
	}
}
