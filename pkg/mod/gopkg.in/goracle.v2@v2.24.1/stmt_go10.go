// +build go1.10

// Copyright 2017 Tamás Gulácsi
//
//
// SPDX-License-Identifier: UPL-1.0 OR Apache-2.0

package goracle

/*
#include <stdlib.h>
#include "dpiImpl.h"

void goracle_setFromString(dpiVar *dv, uint32_t pos, const _GoString_ value) {
	uint32_t length;
	length = _GoStringLen(value);
	if( length == 0 ) {
		return;
	}
	dpiVar_setFromBytes(dv, pos, _GoStringPtr(value), length);
}
*/
import "C"
import (
	//"context"
	"strings"
	"sync"
)

const go10 = true

func dpiSetFromString(dv *C.dpiVar, pos C.uint32_t, x string) {
	C.goracle_setFromString(dv, pos, x)
}

var stringBuilders = stringBuilderPool{
	p: &sync.Pool{New: func() interface{} { return &strings.Builder{} }},
}

type stringBuilderPool struct {
	p *sync.Pool
}

func (sb stringBuilderPool) Get() *strings.Builder {
	return sb.p.Get().(*strings.Builder)
}
func (sb *stringBuilderPool) Put(b *strings.Builder) {
	b.Reset()
	sb.p.Put(b)
}

/*
// ResetSession is called while a connection is in the connection
// pool. No queries will run on this connection until this method returns.
//
// If the connection is bad this should return driver.ErrBadConn to prevent
// the connection from being returned to the connection pool. Any other
// error will be discarded.
func (c *conn) ResetSession(ctx context.Context) error {
	if Log != nil {
		Log("msg", "ResetSession", "conn", c.dpiConn)
	}
	//subCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	//err := c.Ping(subCtx)
	//cancel()
	return c.Ping(ctx)
}
*/
