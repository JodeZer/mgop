package mgop

import (
	"sync"
	_"fmt"
)

type upstreamSessionPool interface {
	appendSession(*sessionWrapper)
	getBest() *sessionWrapper
}

/**
one kind of upstream for polling, just for test
suppose support for insert,update and findAndModify is OK
 */

type pollingSessionPool struct {
	wrappers  []*sessionWrapper
	maxSize   int
	next      int
	nextMutex sync.Mutex
}

func (p *pollingSessionPool)getBest() *sessionWrapper {
	p.nextMutex.Lock()
	p.next = (p.next + 1) % len(p.wrappers)
	cur := p.next
	p.nextMutex.Unlock()
	//fmt.Printf("get %d\n", cur)
	return p.wrappers[cur]
}

func newPollingSessionPool(maxSize int) upstreamSessionPool {
	p := &pollingSessionPool{
		maxSize:maxSize,
		next:-1,
	}
	p.wrappers = make([]*sessionWrapper, 0, maxSize)
	return p
}

func (p *pollingSessionPool)appendSession(sw *sessionWrapper) {
	p.wrappers = append(p.wrappers, sw)
}