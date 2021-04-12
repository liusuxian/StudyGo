package main

import (
    "fmt"
    "math/rand"
    "sync/atomic"
    "unsafe"
)

type Value struct {
    v interface{}
}

type ifaceWords struct {
    typ  unsafe.Pointer
    data unsafe.Pointer
}

type Config struct {
    NodeName string
    Addr     string
    Count    int32
}

func (v *Value) CompareAndSwap(old, new interface{}) (swapped bool) {
    vp := (*ifaceWords)(unsafe.Pointer(v))
    typ := atomic.LoadPointer(&vp.typ)
    oldp := (*ifaceWords)(unsafe.Pointer(&old))
    newp := (*ifaceWords)(unsafe.Pointer(&new))
    fmt.Println(old, new)
    fmt.Println(typ, oldp, newp)
    vv := (*string)(newp.data)
    *vv = "lsx"
    fmt.Println(old, new)
    return false

    //if old.Type != new.Type {
    //    panic
    //}
    //if old.Type != v.Type {
    //    panic
    //}
    //p := v.Data
    //if t.direct {
    //    if p != old.Data {
    //        return false
    //    }
    //} else {
    //    if !old.Type.Equal(p, old.Data) {
    //        return false
    //    } // using type's equality function
    //}
    //return atomic.CompareAndSwapPointer(&v.Data, p, new.Data)
}

func main() {
    configval := &Value{}
    config := &Config{
        NodeName: "北京",
        Addr:     "10.77.95.27",
        Count:    rand.Int31(),
    }
    configval.CompareAndSwap(&Config{}, config)
}
