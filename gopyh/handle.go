// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gopyh provides the variable gopyvar manager for gopy.
// The gopyvars map can NOT be globally shared in C because it
// must use interface{} values that can change location via GC.
// In effect, each gopy package must be thought of as a completely
// separate Go instance, and there can be NO sharing of anything
// between them, because they fundamentally live in different .so
// libraries.
// Thus, we must ensure that all gopyvars used within a package
// are registered within that same package -- this means python
// users typically will import a single package, which exports
// all the relevant functionality as needed.  Any further packages
// cannot share anything other than basic types (int, string etc).
package gopyh

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
)

// GoGopyvar is the type for the gopyvar
type GoGopyvar int64
type CGoGopyvar int64

// --- variable gopyvars: all pointers managed via gopyvars ---

var (
	mu       sync.RWMutex
	ctr      int64
	gopyvars map[GoGopyvar]interface{}
	counts   map[GoGopyvar]int64
)

// IfaceIsNil returns true if interface or value represented by interface is nil
func IfaceIsNil(it interface{}) bool {
	if it == nil {
		return true
	}
	v := reflect.ValueOf(it)
	vk := v.Kind()
	if vk == reflect.Ptr || vk == reflect.Interface || vk == reflect.Map || vk == reflect.Slice || vk == reflect.Func || vk == reflect.Chan {
		return v.IsNil()
	}
	return false
}

// NonPtrValue returns the non-pointer underlying value
func NonPtrValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

// PtrValue returns the pointer version (Addr()) of the underlying value if
// the value is not already a Ptr
func PtrValue(v reflect.Value) reflect.Value {
	if v.CanAddr() && v.Kind() != reflect.Ptr {
		v = v.Addr()
	}
	return v
}

// Embed returns the embedded struct (in first field only) of given type within given struct
func Embed(stru interface{}, embed reflect.Type) interface{} {
	if IfaceIsNil(stru) {
		return nil
	}
	v := NonPtrValue(reflect.ValueOf(stru))
	typ := v.Type()
	if typ == embed {
		return PtrValue(v).Interface()
	}
	if typ.NumField() == 0 {
		return nil
	}
	f := typ.Field(0)
	if f.Type.Kind() == reflect.Struct && f.Anonymous { // anon only avail on StructField fm typ
		vf := v.Field(0)
		vfpi := PtrValue(vf).Interface()
		if f.Type == embed {
			return vfpi
		}
		rv := Embed(vfpi, embed)
		if rv != nil {
			return rv
		}
	}
	return nil
}

var (
	trace = false
)

func init() {
	if len(os.Getenv("GOPY_GOPYVAR_TRACE")) > 0 {
		trace = true
	}
}

// Register registers a new variable instance.
func Register(typnm string, ifc interface{}) CGoGopyvar {
	if IfaceIsNil(ifc) {
		return -1
	}
	mu.Lock()
	defer mu.Unlock()
	if gopyvars == nil {
		gopyvars = make(map[GoGopyvar]interface{})
		counts = make(map[GoGopyvar]int64)
	}
	ctr++
	hc := ctr
	ghc := GoGopyvar(hc)
	gopyvars[ghc] = ifc
	counts[ghc] = 0
	if trace {
		fmt.Printf("gopy Registered: %s %v %d\n", typnm, ifc, hc)
	}
	return CGoGopyvar(hc)
}

// DecRef decrements the reference count for the specified gopyvar
// and removes it if the reference count goes to zero.
func DecRef(gopyvar CGoGopyvar) {
	if gopyvar < 1 {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if gopyvars == nil {
		return
	}
	ghc := GoGopyvar(gopyvar)
	if _, exists := gopyvars[ghc]; !exists {
		return
	}
	counts[ghc]--
	switch cnt := counts[ghc]; {
	case cnt == 0:
		delete(counts, ghc)
		delete(gopyvars, ghc)
		if trace {
			fmt.Printf("gopy DecRef: %d\n", gopyvar)
		}
	case cnt < 0:
		panic(fmt.Sprintf("gopy DecRef ref count %v for gopyvar: %v, ifc %v", cnt, ghc, gopyvars[ghc]))
	default:
		if trace {
			fmt.Printf("gopy DecRef: %d: %d\n", gopyvar, cnt)
		}
	}
}

// IncRef increments the reference count for the specified gopyvar.
func IncRef(gopyvar CGoGopyvar) {
	if gopyvar < 1 {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	ghc := GoGopyvar(gopyvar)
	if _, exists := counts[ghc]; exists {
		counts[ghc]++
		if trace {
			fmt.Printf("gopy IncRef: %d: %d\n", gopyvar, counts[ghc])
		}
	}

}

// VarFromGopyvar gets variable from gopyvar string.
// Reports error to python but does not return it,
// for use in inline calls
func VarFromGopyvar(h CGoGopyvar, typnm string) interface{} {
	v, _ := VarFromGopyvarTry(h, typnm)
	return v
}

// VarFromGopyvarTry version returns the error explicitly,
// for use when error can be processed
func VarFromGopyvarTry(h CGoGopyvar, typnm string) (interface{}, error) {
	if h < 1 {
		return nil, fmt.Errorf("gopy: nil gopyvar")
	}
	mu.RLock()
	defer mu.RUnlock()
	v, has := gopyvars[GoGopyvar(h)]
	if !has {
		err := fmt.Errorf("gopy: variable gopyvar not registered: " + strconv.FormatInt(int64(h), 10))
		// TODO: need to get access to this:
		// C.PyErr_SetString(C.PyExc_TypeError, C.CString(err.Error()))
		return nil, err
	}
	return v, nil
}

// NumGopyvars returns the number of gopyvars in use.
func NumGopyvars() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(gopyvars)
}
