// Copyright 2016 The go-simplechain Authors
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

package network

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	p2ptest "github.com/simplechain-org/go-simplechain/p2p/testing"
	"github.com/simplechain-org/go-simplechain/swarm/state"
)

func newHiveTester(t *testing.T, params *HiveParams, n int, store state.Store) (*bzzTester, *Hive) {
	// setup
	addr := RandomAddr() // tested peers peer address
	to := NewKademlia(addr.OAddr, NewKadParams())
	pp := NewHive(params, to, store) // hive

	return newBzzBaseTester(t, n, addr, DiscoverySpec, pp.Run), pp
}

func TestRegisterAndConnect(t *testing.T) {
	params := NewHiveParams()
	s, pp := newHiveTester(t, params, 1, nil)

	id := s.IDs[0]
	raddr := NewAddrFromNodeID(id)
	pp.Register([]OverlayAddr{OverlayAddr(raddr)})

	// start the hive and wait for the connection
	err := pp.Start(s.Server)
	if err != nil {
		t.Fatal(err)
	}
	defer pp.Stop()
	// retrieve and broadcast
	err = s.TestDisconnected(&p2ptest.Disconnect{
		Peer:  s.IDs[0],
		Error: nil,
	})

	if err == nil || err.Error() != "timed out waiting for peers to disconnect" {
		t.Fatalf("expected peer to connect")
	}
}

func TestHiveStatePersistance(t *testing.T) {
	log.SetOutput(os.Stdout)

	dir, err := ioutil.TempDir("", "hive_test_store")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	store, err := state.NewDBStore(dir) //start the hive with an empty dbstore

	params := NewHiveParams()
	s, pp := newHiveTester(t, params, 5, store)

	peers := make(map[string]bool)
	for _, id := range s.IDs {
		raddr := NewAddrFromNodeID(id)
		pp.Register([]OverlayAddr{OverlayAddr(raddr)})
		peers[raddr.String()] = true
	}

	// start the hive and wait for the connection
	err = pp.Start(s.Server)
	if err != nil {
		t.Fatal(err)
	}
	pp.Stop()
	store.Close()

	persistedStore, err := state.NewDBStore(dir) //start the hive with an empty dbstore

	s1, pp := newHiveTester(t, params, 1, persistedStore)

	//start the hive and wait for the connection

	pp.Start(s1.Server)
	i := 0
	pp.Overlay.EachAddr(nil, 256, func(addr OverlayAddr, po int, nn bool) bool {
		delete(peers, addr.(*BzzAddr).String())
		i++
		return true
	})
	if len(peers) != 0 || i != 5 {
		t.Fatalf("invalid peers loaded")
	}
}
