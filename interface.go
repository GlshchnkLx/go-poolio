package poolio

import "io"

//--------------------------------------------------------------------------------//
// PoolIO Interface
//--------------------------------------------------------------------------------//

// Pool implements a complete interface for working with pool storage
type Pool interface {
	General
	ReadWriteCloser
}

// General implements a general interface for working with pool storage
type General interface {
	New(poolName string, unitAmount int) (Pool, error) // creates a branch of the pool
}

// Reader implements an interface for reading data from a pool
type Reader interface {
	General
	io.Reader
}

// Writer implements an interface for writing data to a pool
type Writer interface {
	General
	io.Writer
}

// Closer implements an interface for closing a pool
type Closer interface {
	General
	io.Closer
}

//--------------------------------------------------------------------------------//

// ReadWriter is the interface that groups the basic Read and Write methods.
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser is the interface that groups the basic Read and Close methods.
type ReadCloser interface {
	Reader
	Closer
}

// WriteCloser is the interface that groups the basic Write and Close methods.
type WriteCloser interface {
	Writer
	Closer
}

// ReadWriteCloser is the interface that groups the basic Read, Write and Close methods.
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

//--------------------------------------------------------------------------------//
