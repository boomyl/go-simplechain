// Copyright 2017 The go-simplechain Authors
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

package storage

import (
	"hash"
)

const (
	BMTHash     = "BMT"
	SHA3Hash    = "SHA3" // http://golang.org/pkg/hash/#Hash
	DefaultHash = BMTHash
)

type SwarmHash interface {
	hash.Hash
	ResetWithLength([]byte)
}

type HashWithLength struct {
	hash.Hash
}

func (h *HashWithLength) ResetWithLength(length []byte) {
	h.Reset()
	h.Write(length)
}
