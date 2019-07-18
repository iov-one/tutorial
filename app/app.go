package app

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/app"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/orm"
	"github.com/iov-one/weave/x"
	"github.com/iov-one/weave/x/msgfee"
	"github.com/iov-one/weave/x/multisig"
	"github.com/iov-one/weave/x/sigs"
	"github.com/iov-one/weave/x/utils"
)

// Authenticator returns authentication with multisigs
// and public key signatues
func Authenticator() x.Authenticator {
	return x.ChainAuth(sigs.Authenticate{}, multisig.Authenticate{})
}

// Chain returns a chain of decorators, to handle authentication,
// fees, logging, and recovery
func Chain(authFn x.Authenticator, minFee coin.Coin) app.Decorators {

	// TODO implement orderbook controller
	return app.ChainDecorators(
		utils.NewLogging(),
		utils.NewRecovery(),
		utils.NewKeyTagger(),
		utils.NewSavepoint().OnCheck(),
		sigs.NewDecorator(),
		multisig.NewDecorator(authFn),
		msgfee.NewFeeDecorator(),
	)
}

// Router returns a default router
func Router(authFn x.Authenticator) app.Router {
	r := app.NewRouter()
	// TODO implement orderbook router
	return r
}

// QueryRouter returns a default query router,
// allowing access to "/auth", "/contracts" and "/"
func QueryRouter() weave.QueryRouter {
	r := weave.NewQueryRouter()
	r.RegisterAll(
		sigs.RegisterQuery,
		multisig.RegisterQuery,
		orm.RegisterQuery,
	)
	return r
}

// Stack wires up a standard router with a standard decorator
// chain. This can be passed into BaseApp.
func Stack(minFee coin.Coin) weave.Handler {
	authFn := Authenticator()
	return Chain(authFn, minFee).WithHandler(Router(authFn))
}
