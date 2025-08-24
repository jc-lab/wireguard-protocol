/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2025 JC-Lab. All Rights Reserved.
 */

package device

type PacketIOEvent int

const (
	PacketIOEventUp PacketIOEvent = 1 << iota
	PacketIOEventDown
	PacketIOEventMTUUpdate
)

type PacketIO interface {
	Read(outputs []OutboundContainer, bufFactory MessageBufferFactory) (int, error)

	// Write writes the provided buffers to the packet device. Offset is the
	// same semantics as the previous tun.Device.Write.
	Write(inputs []InboundContainer) (int, error)

	// MTU returns the current MTU for the device.
	MTU() (int, error)

	// Events returns a channel that yields packet device events.
	// The concrete event type is tun.Event (this adapter bridges to it).
	Events() <-chan PacketIOEvent

	// Close closes the device.
	Close() error

	// BatchSize returns the batch size this device supports.
	BatchSize() int
}
