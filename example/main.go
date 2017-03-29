package main

import (
	"github.com/JodeZer/mgop"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"sync"
)

const connection = `mongodb://admin:JSDBuoydfo76Ykmn3DFo3R@192.168.1.203:27017/admin`

func main() {
	p, err := mgop.DialStrongPool(connection, 5)
	if err != nil {
		fmt.Printf("err !!%s", err)
		return
	}
	sp := sync.WaitGroup{}
	for i:=0;i<1000;i++{
		sp.Add(1)
		go func() {
			sp.Add(1)
			s := p.AcquireSession()
			defer s.Release()
			s.DB("quickpay").C("pt").Insert(bson.M{"iid":i})
			sp.Done()
		}()
		sp.Done()
	}
	sp.Wait()
}
