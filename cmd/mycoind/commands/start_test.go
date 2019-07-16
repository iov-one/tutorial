package commands

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/iov-one/weave/commands/server"
	"github.com/iov-one/weave/tmtest"
	"github.com/iov-one/weave/weavetest/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/iov-one/tutorial/cmd/mycoind/app"
)

func TestStartStandAlone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping ABCI stand-alone test")
	}

	home := setupConfig(t)
	defer os.RemoveAll(home)

	logger := log.NewNopLogger()

	err := server.InitCmd(app.GenInitOptions, logger, home, nil)
	assert.Nil(t, err)

	// set up app and start up
	args := []string{"-bind", "localhost:11122"}
	runStart := func() error {
		return server.StartCmd(app.GenerateApp, logger, home, args)
	}
	timeout := time.Duration(2) * time.Second
	err = runOrTimeout(runStart, timeout)
	assert.Nil(t, err)
}

func runOrTimeout(cmd func() error, timeout time.Duration) error {
	done := make(chan error)
	go func(out chan<- error) {
		// we assume cmd should block (RunForever)
		err := cmd()
		if err != nil {
			out <- err
		}
		out <- errors.New("start died for unknown reasons")
	}(done)

	timer := time.NewTimer(timeout)
	select {
	case err := <-done:
		return err
	case <-timer.C:
		return nil
	}
}

func TestStartWithTendermint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Tendermint integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	home := setupConfig(t)
	defer os.RemoveAll(home)

	const runTime = 5     // how many seconds to run both processes
	const startupTime = 2 // how many seconds to let tendermint startup

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "test-cmd")
	err := server.InitCmd(app.GenInitOptions, logger, home, []string{"-all=f"})
	assert.Nil(t, err)

	tmtest.RunTendermint(ctx, t, home)

	done := make(chan error, 1)
	go func() {
		args := []string{
			"-bind", "localhost:46658",
		}
		done <- server.StartCmd(app.GenerateApp, logger, home, args)
	}()

	select {
	case <-ctx.Done():
		t.Logf("context cancelled before application finished")
	case err := <-done:
		if err != nil {
			t.Fatalf("application failed: %s", err)
		}
	}
}
