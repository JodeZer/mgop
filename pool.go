package mgop

import (
	"gopkg.in/mgo.v2"
	"sync"
	"time"
	"gopkg.in/mgo.v2/bson"
)

type SessionPool interface {
	AcquireSession() *sessionWrapper
	Size() int
}

type StrongSessionPool struct {
	poolmap map[mgo.Mode]upstreamSessionPool
	rw      sync.RWMutex
	//cache a session for ping
	size    int
}

var modeMap map[mgo.Mode]int = map[mgo.Mode]int{
	mgo.Monotonic:1,
	mgo.PrimaryPreferred:1,
	mgo.Secondary:1,
	mgo.SecondaryPreferred:2,
	mgo.Nearest:1,
}

//TODO
func dialMixedPool(url string, fixedSize int) (SessionPool, error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	s.SetMode(mgo.Strong, true)
	s.SetSafe(&mgo.Safe{})

	p := &StrongSessionPool{}
	p.poolmap = make(map[mgo.Mode]upstreamSessionPool, 7)
	for k, v := range modeMap {
		p.poolmap[k] = newPollingSessionPool(v)
		for i := 0; i < v; i++ {
			scp := s.Copy()
			scp.SetMode(k, true)
			p.poolmap[k].appendSession(newSessionWrapper(p, scp))
		}
	}

	p.poolmap[mgo.Strong] = newPollingSessionPool(fixedSize)

	for i := 0; i < fixedSize - 1; i++ {
		scp := s.Copy()
		scp.SetMode(mgo.Strong, true)
		scp.SetSafe(&mgo.Safe{})
		p.poolmap[mgo.Strong].appendSession(newSessionWrapper(p, scp))
	}
	p.poolmap[mgo.Strong].appendSession(newSessionWrapper(p, s))
	return p, nil
}

func DialStrongPool(url string, fixedSize int) (SessionPool, error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	p := &StrongSessionPool{
		size: fixedSize,
	}
	p.poolmap = make(map[mgo.Mode]upstreamSessionPool, 1)

	p.poolmap[mgo.Strong] = newPollingSessionPool(fixedSize)

	for i := 0; i < fixedSize - 1; i++ {
		scp := s.Copy()
		scp.SetMode(mgo.Strong, true)
		p.poolmap[mgo.Strong].appendSession(newSessionWrapper(p, scp))
	}
	s.SetMode(mgo.Strong, true)
	p.poolmap[mgo.Strong].appendSession(newSessionWrapper(p, s))
	go p.pinger()
	return p, nil
}

func (p *StrongSessionPool) AcquireSession() *sessionWrapper {
	p.rw.RLock()
	defer p.rw.RUnlock()
	return p.poolmap[mgo.Strong].getBest()
}

func (p *StrongSessionPool) Size() int {
	return p.size
}

func (p *StrongSessionPool)pinger() {
	for {

		time.Sleep(5 * time.Second)
		p.rw.Lock()
		p.poolmap[mgo.Strong].foreach(func(sw *sessionWrapper) {
			//if session preserved socket is gone or not master,
			//refresh the session
			//TODO ismaster is less quick than ping;may a hash for host will work
			var result isMasterResult
			if err := sw.s.Run("ismaster", &result); err != nil || result.IsMaster {
				sw.refresh()
			}

		}, true)
		p.rw.Unlock()

	}
}

// TODO
func (p *StrongSessionPool) acquireSessionWithMode(mode int) (*sessionWrapper, error) {
	return nil, nil
}

// TODO
func (p *StrongSessionPool) acquireSessionWithFraction(i int32) *sessionWrapper {
	return nil
}


// isMater cmd result
type isMasterResult struct {
	IsMaster   bool `bson:"ismaster"`
	Secondary  bool `bson:"secondary"`
	Primary    string    `bson:"primary"`
	Hosts      []string `bson:"hosts"`
	Me         string  `bson:"me"`
	SetName    string `bson:"setName"`
	ElectionId bson.ObjectId `bson:"electionId"`
}
