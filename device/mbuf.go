/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2025 JC-Lab. All Rights Reserved.
 */

package device

import (
	"sync/atomic"
)

type MessageBufferFactory func() *MessageBuffer

type MessageBuffer struct {
	factory  MessageBufferFactory
	data     [MaxMessageSize]byte
	offset   int
	refcount atomic.Int32

	size int
	cow  bool
}

func (b *MessageBuffer) SetSize(n int, multiuse bool) {
	b.size = n
	b.cow = multiuse
}

func (b *MessageBuffer) Cap() int {
	return len(b.data) - b.offset
}

func (b *MessageBuffer) Len() int {
	return b.size
}

func (b *MessageBuffer) View() []byte {
	return b.data[b.offset:]
}

func (b *MessageBuffer) Retain() *MessageBuffer {
	if b.cow {
		return b.Copy()
	} else {
		b.refcount.Add(1)
		return b
	}
}

func (b *MessageBuffer) Copy() *MessageBuffer {
	newBuf := b.factory()
	newBuf.offset = b.offset
	newBuf.refcount.Store(1)
	newBuf.size = b.size
	newBuf.cow = false
	copy(newBuf.data[:], b.data[:])
	return newBuf
}
