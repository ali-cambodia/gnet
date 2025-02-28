// Copyright 2024 Hello9 Authors
//  All rights reserved.
//
// Author: Hello9
//

package gnet

import (
	"github.com/panjf2000/gnet/v2/internal/queue"
	"github.com/panjf2000/gnet/v2/pkg/errors"
)

// AsyncWrite - AsyncWrite
func (e Engine) AsyncWrite(connId int64, data []byte) error {
	if e.eng == nil {
		return errors.ErrEmptyEngine
	}

	elidx := int(connId >> 48 & 0xffff)
	id := uint16(connId >> 32 & 0xffff)
	fd := int(connId & 0xffffffff)

	e.eng.eventLoops.iterate(func(i int, el *eventloop) bool {
		if i == elidx {
			_ = el.poller.Trigger(queue.HighPriority, func(_ interface{}) error {
				if c := el.connections.getConn(fd); c != nil && c.id == id {
					if !c.opened {
						return nil
					}
					c.write(data)
				}
				return nil
			}, nil)
			return false
		}
		return true
	})

	return nil
}

// Trigger - Trigger
func (e Engine) Trigger(connId int64, cb func(c Conn)) {
	if e.eng == nil {
		return
	}

	if cb == nil {
		return
	}

	elidx := int(connId >> 48 & 0xffff)
	id := uint16(connId >> 32 & 0xffff)
	fd := int(connId & 0xffffffff)

	// (queue.LowPriority, func(_ interface{}
	e.eng.eventLoops.iterate(func(i int, el *eventloop) bool {
		if i == elidx {
			_ = el.poller.Trigger(queue.HighPriority, func(_ interface{}) error {
				if c := el.connections.getConn(fd); c != nil && c.id == id {
					if c.opened {
						cb(c)
					}
				}
				return nil
			}, nil)
			return false
		}
		return true
	})
}

// Iterate - iterate all conns
func (e Engine) Iterate(cb func(c Conn)) {
	if e.Validate() != nil {
		return
	}

	e.eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		_ = el.poller.Trigger(queue.LowPriority, func(_ interface{}) error {
			el.connections.iterate(func(c *conn) bool {
				cb(c)
				return true
			})
			return nil
		}, nil)
		return true
	})
}
