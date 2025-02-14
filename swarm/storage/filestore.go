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

package storage

import (
	"context"
	"io"
)

/*
FileStore provides the client API entrypoints Store and Retrieve to store and retrieve
It can store anything that has a byte slice representation, so files or serialised objects etc.

Storage: FileStore calls the Chunker to segment the input datastream of any size to a merkle hashed tree of chunks. The key of the root block is returned to the client.

Retrieval: given the key of the root block, the FileStore retrieves the block chunks and reconstructs the original data and passes it back as a lazy reader. A lazy reader is a reader with on-demand delayed processing, i.e. the chunks needed to reconstruct a large file are only fetched and processed if that particular part of the document is actually read.

As the chunker produces chunks, FileStore dispatches them to its own chunk store
implementation for storage or retrieval.
*/

const (
	defaultLDBCapacity                = 5000000 // capacity for LevelDB, by default 5*10^6*4096 bytes == 20GB
	defaultCacheCapacity              = 10000   // capacity for in-memory chunks' cache
	defaultChunkRequestsCacheCapacity = 5000000 // capacity for container holding outgoing requests for chunks. should be set to LevelDB capacity
)

type FileStore struct {
	ChunkStore
	hashFunc SwarmHasher
}

type FileStoreParams struct {
	Hash string
}

func NewFileStoreParams() *FileStoreParams {
	return &FileStoreParams{
		Hash: DefaultHash,
	}
}

// for testing locally
func NewLocalFileStore(datadir string, basekey []byte) (*FileStore, error) {
	params := NewDefaultLocalStoreParams()
	params.Init(datadir)
	localStore, err := NewLocalStore(params, nil)
	if err != nil {
		return nil, err
	}
	localStore.Validators = append(localStore.Validators, NewContentAddressValidator(MakeHashFunc(DefaultHash)))
	return NewFileStore(localStore, NewFileStoreParams()), nil
}

func NewFileStore(store ChunkStore, params *FileStoreParams) *FileStore {
	hashFunc := MakeHashFunc(params.Hash)
	return &FileStore{
		ChunkStore: store,
		hashFunc:   hashFunc,
	}
}

// Public API. Main entry point for document retrieval directly. Used by the
// FS-aware API and httpaccess
// Chunk retrieval blocks on netStore requests with a timeout so reader will
// report error if retrieval of chunks within requested range time out.
// It returns a reader with the chunk data and whether the content was encrypted
func (f *FileStore) Retrieve(ctx context.Context, addr Address) (reader *LazyChunkReader, isEncrypted bool) {
	isEncrypted = len(addr) > f.hashFunc().Size()
	getter := NewHasherStore(f.ChunkStore, f.hashFunc, isEncrypted)
	reader = TreeJoin(ctx, addr, getter, 0)
	return
}

// Public API. Main entry point for document storage directly. Used by the
// FS-aware API and httpaccess
func (f *FileStore) Store(ctx context.Context, data io.Reader, size int64, toEncrypt bool) (addr Address, wait func(context.Context) error, err error) {
	putter := NewHasherStore(f.ChunkStore, f.hashFunc, toEncrypt)
	return PyramidSplit(ctx, data, putter, putter)
}

func (f *FileStore) HashSize() int {
	return f.hashFunc().Size()
}
