// Copyright 2018 The go-simplechain Authors
// This file is part of the go-simplechain library.
//
// The go-simplechain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-simplechain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-simplechain library. If not, see <http://www.gnu.org/licenses/>.

package simulation

import (
	"context"
	"errors"
	"flag"
	"sync"
	"testing"
	"time"

	colorable "github.com/mattn/go-colorable"
	"github.com/simplechain-org/go-simplechain/log"
	"github.com/simplechain-org/go-simplechain/node"
	"github.com/simplechain-org/go-simplechain/p2p"
	"github.com/simplechain-org/go-simplechain/p2p/simulations/adapters"
	"github.com/simplechain-org/go-simplechain/rpc"
)

var (
	loglevel = flag.Int("loglevel", 2, "verbosity of logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

// TestRun tests if Run method calls RunFunc and if it handles context properly.
func TestRun(t *testing.T) {
	sim := New(noopServiceFuncMap)
	defer sim.Close()

	t.Run("call", func(t *testing.T) {
		expect := "something"
		var got string
		r := sim.Run(context.Background(), func(ctx context.Context, sim *Simulation) error {
			got = expect
			return nil
		})

		if r.Error != nil {
			t.Errorf("unexpected error: %v", r.Error)
		}
		if got != expect {
			t.Errorf("expected %q, got %q", expect, got)
		}
	})

	t.Run("cancelation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		r := sim.Run(ctx, func(ctx context.Context, sim *Simulation) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})

		if r.Error != context.DeadlineExceeded {
			t.Errorf("unexpected error: %v", r.Error)
		}
	})

	t.Run("context value and duration", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "hey", "there")
		sleep := 50 * time.Millisecond

		r := sim.Run(ctx, func(ctx context.Context, sim *Simulation) error {
			if ctx.Value("hey") != "there" {
				return errors.New("expected context value not passed")
			}
			time.Sleep(sleep)
			return nil
		})

		if r.Error != nil {
			t.Errorf("unexpected error: %v", r.Error)
		}
		if r.Duration < sleep {
			t.Errorf("reported run duration less then expected: %s", r.Duration)
		}
	})
}

// TestClose tests are Close method triggers all close functions and are all nodes not up anymore.
func TestClose(t *testing.T) {
	var mu sync.Mutex
	var cleanupCount int

	sleep := 50 * time.Millisecond

	sim := New(map[string]ServiceFunc{
		"noop": func(ctx *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			return newNoopService(), func() {
				time.Sleep(sleep)
				mu.Lock()
				defer mu.Unlock()
				cleanupCount++
			}, nil
		},
	})

	nodeCount := 30

	_, err := sim.AddNodes(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	var upNodeCount int
	for _, n := range sim.Net.GetNodes() {
		if n.Up {
			upNodeCount++
		}
	}
	if upNodeCount != nodeCount {
		t.Errorf("all nodes should be up, insted only %v are up", upNodeCount)
	}

	sim.Close()

	if cleanupCount != nodeCount {
		t.Errorf("number of cleanups expected %v, got %v", nodeCount, cleanupCount)
	}

	upNodeCount = 0
	for _, n := range sim.Net.GetNodes() {
		if n.Up {
			upNodeCount++
		}
	}
	if upNodeCount != 0 {
		t.Errorf("all nodes should be down, insted %v are up", upNodeCount)
	}
}

// TestDone checks if Close method triggers the closing of done channel.
func TestDone(t *testing.T) {
	sim := New(noopServiceFuncMap)
	sleep := 50 * time.Millisecond
	timeout := 2 * time.Second

	start := time.Now()
	go func() {
		time.Sleep(sleep)
		sim.Close()
	}()

	select {
	case <-time.After(timeout):
		t.Error("done channel closing timmed out")
	case <-sim.Done():
		if d := time.Since(start); d < sleep {
			t.Errorf("done channel closed sooner then expected: %s", d)
		}
	}
}

// a helper map for usual services that do not do anyting
var noopServiceFuncMap = map[string]ServiceFunc{
	"noop": noopServiceFunc,
}

// a helper function for most basic noop service
func noopServiceFunc(ctx *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
	return newNoopService(), nil, nil
}

// noopService is the service that does not do anything
// but implements node.Service interface.
type noopService struct{}

func newNoopService() node.Service {
	return &noopService{}
}

func (t *noopService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (t *noopService) APIs() []rpc.API {
	return []rpc.API{}
}

func (t *noopService) Start(server *p2p.Server) error {
	return nil
}

func (t *noopService) Stop() error {
	return nil
}
