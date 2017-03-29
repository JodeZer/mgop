package mgop

import (
	"gopkg.in/mgo.v2"
	"sync"
)

type SessionPool interface {
	AcquireSession() *sessionWrapper
}

type StrongSessionPool struct {
	m  map[mgo.Mode]upstreamSessionPool
	rw sync.RWMutex
}

var modeMap map[mgo.Mode]int = map[mgo.Mode]int{
	mgo.Monotonic:1,
	mgo.PrimaryPreferred:1,
	mgo.Secondary:1,
	mgo.SecondaryPreferred:2,
	mgo.Nearest:1,
}

func DialMixedPool(url string, fixedSize int) (SessionPool, error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	p := &StrongSessionPool{}
	p.m = make(map[mgo.Mode]upstreamSessionPool, 7)
	for k, v := range modeMap {
		p.m[k] = newPollingSessionPool(v)
		for i := 0; i < v; i++ {
			scp := s.Copy()
			scp.SetMode(k, true)
			p.m[k].appendSession(newSessionWrapper(p, scp))
		}
	}

	p.m[mgo.Strong] = newPollingSessionPool(fixedSize)

	for i := 0; i < fixedSize - 1; i++ {
		scp := s.Copy()
		scp.SetMode(mgo.Strong, true)
		p.m[mgo.Strong].appendSession(newSessionWrapper(p, scp))
	}
	s.SetMode(mgo.Strong, true)
	p.m[mgo.Strong].appendSession(newSessionWrapper(p, s))
	return p, nil
}

func DialStrongPool(url string, fixedSize int) (SessionPool, error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	p := &StrongSessionPool{}
	p.m = make(map[mgo.Mode]upstreamSessionPool, 1)

	p.m[mgo.Strong] = newPollingSessionPool(fixedSize)

	for i := 0; i < fixedSize - 1; i++ {
		scp := s.Copy()
		scp.SetMode(mgo.Strong, true)
		p.m[mgo.Strong].appendSession(newSessionWrapper(p, scp))
	}
	s.SetMode(mgo.Strong, true)
	p.m[mgo.Strong].appendSession(newSessionWrapper(p, s))
	return p, nil
}

func (p *StrongSessionPool) AcquireSession() *sessionWrapper {
	p.rw.RLock()
	defer p.rw.RUnlock()
	return p.m[mgo.Strong].getBest()
}

// TODO
func (p *StrongSessionPool) acquireSessionWithMode(mode int) (*sessionWrapper, error) {
	return nil, nil
}

// TODO
func (p *StrongSessionPool) acquireSessionWithFraction(i int32) *sessionWrapper {
	return nil
}
