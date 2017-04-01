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
	host     string // which host the session is connect to
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

// if a reset happen ,return true else return false
func (sw *sessionWrapper)refreshIfNotMaster() bool {
	result := &isMasterResult{}
	sw.s.Run("ismaster", result)
	if result.IsMaster {
		return false
	}
	sw.s.Refresh()
	return true
}

func (sw *sessionWrapper)refresh() {
	sw.s.Refresh()
	sw.s.Ping()
}

func newSessionWrapper(p SessionPool, session *mgo.Session) *sessionWrapper {

	sess := &sessionWrapper{
		s :session,
		belongTo:p,
		ref:0,
	}
	return sess
}
