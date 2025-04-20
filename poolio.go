package poolio

import (
	"errors"
)

//--------------------------------------------------------------------------------//
// PoolIO: Const
//--------------------------------------------------------------------------------//

var (
	ErrPoolIsClosed                    = errors.New("pool is closed")
	ErrPoolStorageSizeLessZero         = errors.New("pool storage size must be greater than zero")
	ErrPoolBranchStorageSizeMoreParent = errors.New("storage size of the pool branch must be less than or equal to the parent pool")
)

//--------------------------------------------------------------------------------//
// PoolIO: Object method
//--------------------------------------------------------------------------------//

type _Pool struct {
	mutex  chan interface{}
	name   string
	parent *_Pool
	branch map[string]*_Pool

	unitLength int
	unitAmount int

	storageLength   int
	storageArray    []byte
	storageChan     chan []byte
	storageChanRead chan []byte

	isClosed bool
}

// mutex lock / unlock
func (pool *_Pool) lock() {
	pool.mutex <- true
}

func (pool *_Pool) unlock() {
	<-pool.mutex
}

// Reader implements
func (pool *_Pool) Read(p []byte) (n int, err error) {
	unitValue, unitCheck := <-pool.storageChanRead
	if !unitCheck {
		return 0, ErrPoolIsClosed
	}

	n = copy(p, unitValue)
	pool.storageChan <- unitValue[:pool.unitLength]

	return
}

// Writer implements
func (pool *_Pool) Write(p []byte) (n int, err error) {
	pool.lock()
	defer pool.unlock()

	if pool.isClosed {
		return 0, ErrPoolIsClosed
	}

	unitValue := <-pool.storageChan
	n = copy(unitValue, p)
	pool.storageChanRead <- unitValue[:n]

	return
}

// Closer implements
func (pool *_Pool) close() error {
	pool.lock()
	defer pool.unlock()

	if pool.isClosed {
		return ErrPoolIsClosed
	}

	for poolBranchName, poolBranchUnit := range pool.branch {
		err := poolBranchUnit.close()

		if err == nil {
			delete(pool.branch, poolBranchName)
		}
	}

	for done := false; !done; {
		select {
		case unitValue, unitCheck := <-pool.storageChanRead:
			if unitCheck {
				pool.storageChan <- unitValue
			}
		default:
			done = true
		}
	}

	pool.parent = nil
	pool.isClosed = true
	close(pool.storageChanRead)

	<-pool.mutex
	return nil
}

func (pool *_Pool) Close() error {
	pool.lock()
	poolParentUnit := pool.parent
	pool.unlock()

	if poolParentUnit != nil {
		poolParentUnit.mutex <- true
	}

	err := pool.close()
	if err != nil {
		return err
	}

	if poolParentUnit != nil {
		delete(poolParentUnit.branch, pool.name)
		poolParentUnit.mutex <- true
	} else {
		pool.storageArray = nil
		close(pool.storageChan)
	}

	return nil
}

// New implements Pool
func (pool *_Pool) New(poolName string, unitAmount int) (newpool Pool, err error) {
	var (
		storageLength int
	)

	if unitAmount < 0 {
		return nil, ErrPoolStorageSizeLessZero
	}

	if pool.unitAmount < unitAmount {
		return nil, ErrPoolBranchStorageSizeMoreParent
	}

	storageLength = pool.unitLength * unitAmount

	newpool = &_Pool{
		mutex:           make(chan interface{}, 1),
		name:            poolName,
		parent:          pool,
		branch:          map[string]*_Pool{},
		unitLength:      pool.unitLength,
		unitAmount:      unitAmount,
		storageLength:   storageLength,
		storageChan:     pool.storageChan,
		storageChanRead: make(chan []byte, unitAmount),
		isClosed:        false,
	}

	pool.branch[poolName] = newpool.(*_Pool)

	return
}

//--------------------------------------------------------------------------------//
// PoolIO: Package method
//--------------------------------------------------------------------------------//

func New(unitLength, unitAmount int) (Pool, error) {
	var (
		pool *_Pool
		err  error
	)

	var (
		storageLength = unitLength * unitAmount
		storageArray  []byte
	)

	if storageLength <= 0 {
		return nil, ErrPoolStorageSizeLessZero
	}

	storageArray = make([]byte, storageLength)

	pool = &_Pool{
		mutex:           make(chan interface{}, 1),
		name:            "root",
		parent:          nil,
		branch:          map[string]*_Pool{},
		unitLength:      unitLength,
		unitAmount:      unitAmount,
		storageLength:   storageLength,
		storageArray:    storageArray,
		storageChan:     make(chan []byte, unitAmount),
		storageChanRead: make(chan []byte, unitAmount),
		isClosed:        false,
	}

	for i := 0; i < pool.unitAmount; i++ {
		pool.storageChan <- pool.storageArray[i*pool.unitLength : (i+1)*pool.unitLength]
	}

	return pool, err
}

//--------------------------------------------------------------------------------//
