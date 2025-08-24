package device

import (
	"fmt"
	"io"
	"net/netip"
	"os"
	"testing"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/conn/bindtest"
)

// genTestPair creates a testPair.
type testPacketIO struct {
	Inbound  chan []byte // incoming packets, closed on TUN close
	Outbound chan []byte // outbound packets, blocks forever on TUN close

	closed chan struct{}
	events chan PacketIOEvent

	targetKey NoisePublicKey
}

var _ PacketIO = (*testPacketIO)(nil)

func newTestPacketIO() *testPacketIO {
	c := &testPacketIO{
		Inbound:  make(chan []byte),
		Outbound: make(chan []byte),
		closed:   make(chan struct{}),
		events:   make(chan PacketIOEvent, 1),
	}
	c.events <- PacketIOEventUp
	return c
}

func (tpi *testPacketIO) Read(outputs []OutboundContainer, bufFactory MessageBufferFactory) (int, error) {
	// Wait for either an outbound packet or the events channel closing.
	for {
		select {
		case msg, ok := <-tpi.Outbound:
			if !ok {
				return 0, os.ErrClosed
			}
			outputs[0].Broadcast = false
			outputs[0].Peers = []NoisePublicKey{tpi.targetKey}
			outputs[0].Buf = bufFactory()
			n := copy(outputs[0].Buf.View(), msg)
			outputs[0].Buf.SetSize(n, false)
			return 1, nil
		case ev, ok := <-tpi.events:
			_ = ev
			if !ok {
				// events channel closed => device is closed
				return 0, os.ErrClosed
			}
			// otherwise loop and wait for outbound packet
		}
	}
}

func (tpi *testPacketIO) Write(inputs []InboundContainer) (int, error) {
	if len(inputs) == 0 {
		close(tpi.closed)
		close(tpi.events)
		return 0, io.EOF
	}
	for i, input := range inputs {
		msg := make([]byte, input.Buf.Len())
		copy(msg, input.Buf.View())
		select {
		case <-tpi.closed:
			return i, os.ErrClosed
		case tpi.Inbound <- msg:
		}
	}
	return len(inputs), nil
}
func (tpi *testPacketIO) MTU() (int, error) {
	return DefaultMTU, nil
}
func (tpi *testPacketIO) Events() <-chan PacketIOEvent { return tpi.events }
func (tpi *testPacketIO) Close() error {
	tpi.Write(nil)
	return nil
}
func (tpi *testPacketIO) BatchSize() int { return 1 }

// genTestPair creates a testPair.
func genTestPair(tb testing.TB, realSocket bool) (pair testPair) {
	cfg, endpointCfg := genConfigs(tb)
	var binds [2]conn.Bind
	if realSocket {
		binds[0], binds[1] = conn.NewDefaultBind(), conn.NewDefaultBind()
	} else {
		binds = bindtest.NewChannelBinds()
	}

	var ioDevices [2]*testPacketIO
	for i := range ioDevices {
		ioDevices[i] = newTestPacketIO()
	}

	// Bring up a ChannelTun for each config.
	for i := range pair {
		p := &pair[i]
		p.io = ioDevices[i]
		p.ip = netip.AddrFrom4([4]byte{1, 0, 0, byte(i + 1)})
		level := LogLevelVerbose
		if _, ok := tb.(*testing.B); ok && !testing.Verbose() {
			level = LogLevelError
		}
		p.dev = NewDevice(p.io, binds[i], NewLogger(level, fmt.Sprintf("dev%d: ", i)))
		if err := p.dev.IpcSet(cfg[i]); err != nil {
			tb.Errorf("failed to configure device %d: %v", i, err)
			p.dev.Close()
			continue
		}
		ioDevices[i^1].targetKey = p.dev.staticIdentity.publicKey
		if err := p.dev.Up(); err != nil {
			tb.Errorf("failed to bring up device %d: %v", i, err)
			p.dev.Close()
			continue
		}
		endpointCfg[i^1] = fmt.Sprintf(endpointCfg[i^1], p.dev.net.port)
	}
	for i := range pair {
		p := &pair[i]
		if err := p.dev.IpcSet(endpointCfg[i]); err != nil {
			tb.Errorf("failed to configure device endpoint %d: %v", i, err)
			p.dev.Close()
			continue
		}
		// The device is ready. Close it when the test completes.
		tb.Cleanup(p.dev.Close)
	}
	return
}
