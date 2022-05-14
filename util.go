package resource

import (
	"sync/atomic"
)

type flag int32

func (f *flag) check() bool {
	return atomic.LoadInt32((*int32)(f)) != 0
}

func (f *flag) set() {
	atomic.StoreInt32((*int32)(f), 1)
}
