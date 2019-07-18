package app

import (	
	"github.com/iov-one/weave/x"
	"github.com/iov-one/weave/x/sigs"
	"github.com/iov-one/weave/x/multisig"
)

// Authenticator returns authentication with multisigs 
// and public key signatues 
func Authenticator() x.Authenticator {
	return x.ChainAuth(sigs.Authenticate{}, multisig.Authenticate{})
}