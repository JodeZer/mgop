package mgop

import (
	"gopkg.in/mgo.v2"

	"sync/atomic"
)

type sessionWrapper struct {
	s        *mgo.Session
	belongTo SessionPool
	ref      int32
	maxRef   int32
}

func (sw *sessionWrapper) Release() {
	sw.ReleaseWithFraction(-1)
}

func (sw *sessionWrapper) ReleaseWithFraction(i int32) {
	atomic.AddInt32(&sw.ref, i)
}

func (sw *sessionWrapper) atomicAcquire() bool {
	atomic.AddInt32(&sw.ref, 1)
	return true
}

// only allowed native session operation
func (sw *sessionWrapper)DB(name string) *mgo.Database {
	return sw.s.DB(name)
}

func newSessionWrapper(p SessionPool, session *mgo.Session) *sessionWrapper {
	return &sessionWrapper{
		s :session,
		belongTo:p,
		ref:0,
	}
}
