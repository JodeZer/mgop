# mgop<br> [![GoDoc](https://godoc.org/github.com/JodeZer/mgop?status.svg)](https://godoc.org/github.com/JodeZer/mgop)

a mgo session pool
still need more work...

```
go get gopkg.in/mgo.v2
go get github.com/JodeZer/mgop
```

with mgo, client hosts may create too much connections by using Copy method in high concurrency situation,like this

```go
func foo(){
  s := globalSession.Copy()
  defer s.Close()
}
```
especially when using mgo.Eventual mode.

mgop helps mongo doing better in high concurrency situation by buffering some sockets and tranporting requests through these sockets.

```go
func foo(){
  p, _ := mgop.DialPool("127.0.0.1:27017", 5, 5)
  session := p.AcquireSession()
  defer session.Release()
  session.DB("test").C("test").Insert(bson.M{"id":1})
}
```

more deatils seeing example folder
