// mautrix-whatsapp - A Matrix-WhatsApp puppeting bridge.
// Copyright (C) 2026 Kaindl Network
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package connector

import (
	"strings"
	"testing"
	"time"

	"go.mau.fi/whatsmeow/types"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/event"
)

func assertEqual[T comparable](t *testing.T, expected, actual T, msg ...string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v %v", expected, actual, msg)
	}
}

func TestCallTrackerLifecycle(t *testing.T) {
	var ct callTracker
	portal := networkid.PortalKey{ID: "123@lid", Receiver: "login"}
	otherPortal := networkid.PortalKey{ID: "456@lid", Receiver: "login"}
	now := time.Now()
	assertEqual(t, true, ct.add(&trackedCall{ID: "a", Portal: portal, State: callStateRinging, OfferedAt: now}))
	assertEqual(t, false, ct.add(&trackedCall{ID: "a", Portal: portal, State: callStateRinging, OfferedAt: now}), "duplicate offers must be ignored")
	assertEqual(t, true, ct.add(&trackedCall{ID: "b", Portal: otherPortal, State: callStateRinging, OfferedAt: now}))

	assertEqual(t, 1, len(ct.ringingInPortal(portal)))

	accepted := ct.transition("a", func(c *trackedCall) bool {
		c.State = callStateAccepted
		c.AcceptedAt = now
		return true
	})
	assertEqual(t, true, accepted != nil)
	assertEqual(t, 0, len(ct.ringingInPortal(portal)))

	ended := ct.transition("a", func(c *trackedCall) bool {
		c.State = callStateEnded
		c.EndedAt = now.Add(95 * time.Second)
		return true
	})
	assertEqual(t, "📞 Voice call ended (duration 1:35).", ended.text())
	assertEqual(t, true, ct.transition("a", func(c *trackedCall) bool { return true }) == nil, "finished calls must be removed")

	assertEqual(t, true, ct.transition("b", func(c *trackedCall) bool { return false }) == nil, "no-op transitions return nil")
	assertEqual(t, 1, len(ct.ringingInPortal(otherPortal)))
}

func TestCallTrackerExpiry(t *testing.T) {
	var ct callTracker
	ct.add(&trackedCall{ID: "old", State: callStateRinging, OfferedAt: time.Now().Add(-2 * callTrackerMaxAge)})
	ct.add(&trackedCall{ID: "new", State: callStateRinging, OfferedAt: time.Now()})
	_, hasOld := ct.calls["old"]
	assertEqual(t, false, hasOld)
	_, hasNew := ct.calls["new"]
	assertEqual(t, true, hasNew)
}

func TestTrackedCallContent(t *testing.T) {
	call := &trackedCall{Chat: types.NewJID("123", types.HiddenUserServer), Video: true, State: callStateRinging}
	content := call.content()
	assertEqual(t, event.MsgNotice, content.MsgType)
	assertEqual(t, event.BeeperActionMessageCallTypeVideo, content.BeeperActionMessage.CallType)
	assertEqual(t, true, strings.Contains(content.Body, "Incoming video call"))

	call.State = callStateMissed
	assertEqual(t, "📵 Missed video call.", call.text())

	call.Group, call.Video = true, false
	call.State = callStateDeclinedFromMatrix
	assertEqual(t, "📵 Declined group voice call from Matrix.", call.text())
}

func TestFormatCallDuration(t *testing.T) {
	assertEqual(t, "0:05", formatCallDuration(5*time.Second))
	assertEqual(t, "12:00", formatCallDuration(12*time.Minute))
	assertEqual(t, "1:02:03", formatCallDuration(time.Hour+2*time.Minute+3*time.Second))
}
