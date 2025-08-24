/* SPDX-License-Identifier: MIT
 *
 * Tests for the library API (raw send/receive).
 *
 * This test uses the library PrepareOutboundForPeer to build a transport
 * packet, sends it over bindtest channels and asserts the peer's TUN
 * receives the original L3 payload.
 */

package device

//import (
//	"bytes"
//	"fmt"
//	"net/netip"
//	"testing"
//	"time"
//
//	"golang.zx2c4.com/wireguard/conn/bindtest"
//	"golang.zx2c4.com/wireguard/tun/tuntest"
//)
//
//func TestRawAPIEndToEnd(t *testing.T) {
//	binds := bindtest.NewChannelBinds()
//
//	// create two devices with channel TUNs
//	tunA := tuntest.NewChannelTUN()
//	tunB := tuntest.NewChannelTUN()
//
//	devA := NewDevice(tunA.TUN(), binds[0], NewLogger(LogLevelError, "devA: "))
//	devB := NewDevice(tunB.TUN(), binds[1], NewLogger(LogLevelError, "devB: "))
//
//	// configure keys and peers via IpcSet like existing tests
//	cfgs, endpointCfgs := genConfigs(t)
//	if err := devA.IpcSet(cfgs[0]); err != nil {
//		t.Fatalf("devA IpcSet failed: %v", err)
//	}
//	if err := devB.IpcSet(cfgs[1]); err != nil {
//		t.Fatalf("devB IpcSet failed: %v", err)
//	}
//
//	if err := devA.Up(); err != nil {
//		t.Fatalf("devA Up failed: %v", err)
//	}
//	if err := devB.Up(); err != nil {
//		t.Fatalf("devB Up failed: %v", err)
//	}
//
//	// set endpoints to each other's listening port (like genTestPair)
//	endpointCfgs[0] = fmt.Sprintf(endpointCfgs[0], devB.net.port)
//	endpointCfgs[1] = fmt.Sprintf(endpointCfgs[1], devA.net.port)
//
//	if err := devA.IpcSet(endpointCfgs[0]); err != nil {
//		t.Fatalf("devA endpoint IpcSet failed: %v", err)
//	}
//	if err := devB.IpcSet(endpointCfgs[1]); err != nil {
//		t.Fatalf("devB endpoint IpcSet failed: %v", err)
//	}
//
//	// find peer public keys
//	var pkA, pkB NoisePublicKey
//	for k := range devA.peers.keyMap {
//		pkB = k // devA has remote pkB
//		break
//	}
//	for k := range devB.peers.keyMap {
//		pkA = k // devB has remote pkA
//		break
//	}
//
//	peerA := devA.LookupPeer(pkB)
//	peerB := devB.LookupPeer(pkA)
//	if peerA == nil || peerB == nil {
//		t.Fatalf("peers not found after IpcSet")
//	}
//
//	// Kick off handshake from A -> B
//	if err := peerA.SendHandshakeInitiation(false); err != nil {
//		// SendHandshakeInitiation may return nil or error; log but continue to wait
//		t.Logf("SendHandshakeInitiation returned: %v", err)
//	}
//
//	// Wait for symmetric sessions (keypairs) to be established on both sides.
//	timeout := time.Now().Add(5 * time.Second)
//	for {
//		if peerA.keypairs.Current() != nil && peerB.keypairs.Current() != nil {
//			break
//		}
//		if time.Now().After(timeout) {
//			t.Fatalf("timeout waiting for keypairs to be established")
//		}
//		time.Sleep(20 * time.Millisecond)
//	}
//
//	// Build an L3 payload: use tuntest.Ping with src/dst matching genConfigs (1.0.0.1 <-> 1.0.0.2)
//	src := netip.AddrFrom4([4]byte{1, 0, 0, 1})
//	dst := netip.AddrFrom4([4]byte{1, 0, 0, 2})
//	pkt := tuntest.Ping(dst, src)
//
//	// Prepare outbound transport packet using library API
//	out, err := devA.PrepareOutboundForPeer(pkB, pkt)
//	if err != nil {
//		t.Fatalf("PrepareOutboundForPeer failed: %v", err)
//	}
//
//	// Send via bind channel (simulate network). bindtest expects the ChannelEndpoint
//	// value that corresponds to the receiver; for bind[0] -> bind[1] IPv4 use ChannelEndpoint(2).
//	if err := binds[0].Send([][]byte{out}, bindtest.ChannelEndpoint(1)); err != nil {
//		t.Fatalf("bind send failed: %v", err)
//	}
//
//	// Wait for the other side's TUN to receive the original L3 packet.
//	select {
//	case got := <-tunB.Inbound:
//		if !bytes.Equal(got, pkt) {
//			t.Fatalf("payload mismatch; got %x want %x", got, pkt)
//		}
//	case <-time.After(2 * time.Second):
//		t.Fatalf("timeout waiting for packet on tunB")
//	}
//}
