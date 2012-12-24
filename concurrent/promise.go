// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concurrent

import (
	"code.google.com/p/biogo/bio"
	"sync"
)

// Implementation of a promise multiple goroutine synchronisation and communication system
// based on the approach used in Alice. Promises will safely allow multiple promisers to
// interact with multiple promisees.
//
// New or non-error Broken Promises can be Fulfilled or Failed. Fulfilled or Failed Promises
// can be Broken and any state of Promise can be Recovered if specified at creation.
//
// Promises can be mutable or not, recoverable or not and may relay internal error states
// to other listeners.
// Mutable promises may have their value state changed with subsequence Fulfill calls.
// Recoverable promises may be recovered after a Fail call.
// Promises created with relay set to true will relay an error generated by attempting to
// fulfill an immutable fulfilled promise.
type Promise struct {
	message     chan Result
	m           sync.Mutex
	mutable     bool
	recoverable bool
	relay       bool
}

// Create a new promise
func NewPromise(mutable, recoverable, relay bool) *Promise {
	return &Promise{
		message:     make(chan Result, 1),
		mutable:     mutable,
		recoverable: recoverable,
		relay:       relay,
	}

}

func (p *Promise) messageState() (message Result, set bool) {
	select {
	case message = <-p.message:
		set = true
	default:
	}

	return
}

// Fulfill a promise, allowing listeners to unblock.
func (p *Promise) Fulfill(value interface{}) error {
	p.m.Lock()
	defer p.m.Unlock()

	return p.fulfill(value)
}

func (p *Promise) fulfill(value interface{}) (err error) {
	r, set := p.messageState()

	if r.Err != nil {
		err = bio.NewError("Tried to fulfill a failed promise", 0, r.Err)
	} else {
		if !set || p.mutable {
			r.Value = value
			err = nil
		} else {
			err = bio.NewError("Tried to fulfill an already set immutable promise", 0)
		}
	}

	if err != nil && p.relay {
		if r.Err != nil {
			err = bio.NewError("Promise already failed - cannot relay", 0, r.Err)
		} else {
			r.Err = err
		}
	}

	p.message <- r

	return
}

// Fail a promise allowing listeners to unblock, but sending an error state.
func (p *Promise) Fail(value interface{}, err error) (ok bool) {
	p.m.Lock()
	defer p.m.Unlock()

	return p.fail(value, err)
}

func (p *Promise) fail(value interface{}, err error) (f bool) {
	r, _ := p.messageState()

	if r.Err == nil && r.Value == nil {
		if value != nil {
			r.Value = value
		}
		r.Err = err
		f = true
	} else {
		f = false
	}

	p.message <- r

	return
}

// Recover a failed promise, setting the error state to nil. Promise must be recoverable.
func (p *Promise) Recover(value interface{}) (ok bool) {
	p.m.Lock()
	defer p.m.Unlock()

	r, _ := p.messageState()

	if p.recoverable {
		r.Err = nil
		if value != nil {
			p.fulfill(value)
		}
		ok = true
	} else {
		ok = false
	}

	return
}

// Break an already fulfilled or failed promise, blocking all listeners.
func (p *Promise) Break() {
	p.m.Lock()
	defer p.m.Unlock()

	p.messageState()
}

// Wait for a promise to be fulfilled, failed or recovered.
func (p *Promise) Wait() <-chan Result {
	r := <-p.message
	p.message <- r
	f := make(chan Result, 1)
	f <- r
	close(f)
	return f
}
