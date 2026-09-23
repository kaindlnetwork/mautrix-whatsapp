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
	"context"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"
	"maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-whatsapp/pkg/waid"
)

// callEventMaxAge is the maximum age of a call offer that is still bridged.
// Older offers are usually replayed after the bridge was offline and are no longer ringing.
const callEventMaxAge = 15 * time.Minute

// callTrackerMaxAge is how long a call is kept in memory without receiving a final state.
const callTrackerMaxAge = 6 * time.Hour

type callState int

const (
	callStateRinging callState = iota
	callStateAccepted
	callStateMissed
	callStateDeclined
	callStateDeclinedFromMatrix
	callStateEnded
)

// trackedCall is a WhatsApp call for which a notice was sent to Matrix.
// The notice is edited whenever the call state changes, so the Matrix timeline mirrors the WhatsApp call log.
type trackedCall struct {
	ID         string
	Chat       types.JID
	Portal     networkid.PortalKey
	Creator    types.JID
	RejectTo   types.JID // the call creator JID as sent by WhatsApp, used when declining the call
	Video      bool
	Group      bool
	State      callState
	OfferedAt  time.Time
	AcceptedAt time.Time
	EndedAt    time.Time
	MessageID  networkid.MessageID
}

func (tc *trackedCall) callType() event.BeeperActionMessageCallType {
	if tc.Video {
		return event.BeeperActionMessageCallTypeVideo
	}
	return event.BeeperActionMessageCallTypeVoice
}

func (tc *trackedCall) text() string {
	kind := "voice"
	if tc.Video {
		kind = "video"
	}
	if tc.Group {
		kind = "group " + kind
	}
	switch tc.State {
	case callStateRinging:
		return fmt.Sprintf("📞 Incoming %s call. Answer it in the WhatsApp app or decline it with `!wa decline-call`.", kind)
	case callStateAccepted:
		return fmt.Sprintf("📞 %s call in progress (answered on another device).", capitalize(kind))
	case callStateMissed:
		return fmt.Sprintf("📵 Missed %s call.", kind)
	case callStateDeclined:
		return fmt.Sprintf("📵 Declined %s call.", kind)
	case callStateDeclinedFromMatrix:
		return fmt.Sprintf("📵 Declined %s call from Matrix.", kind)
	case callStateEnded:
		if !tc.AcceptedAt.IsZero() && tc.EndedAt.After(tc.AcceptedAt) {
			return fmt.Sprintf("📞 %s call ended (duration %s).", capitalize(kind), formatCallDuration(tc.EndedAt.Sub(tc.AcceptedAt)))
		}
		return fmt.Sprintf("📞 %s call ended.", capitalize(kind))
	default:
		return fmt.Sprintf("📞 %s call.", capitalize(kind))
	}
}

func (tc *trackedCall) content() *event.MessageEventContent {
	return &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    tc.text(),
		BeeperActionMessage: &event.BeeperActionMessage{
			Type:     event.BeeperActionMessageCall,
			CallType: tc.callType(),
		},
	}
}

func capitalize(s string) string {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return s
	}
	return string(s[0]-'a'+'A') + s[1:]
}

func formatCallDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d / time.Hour)
	m := int(d/time.Minute) % 60
	s := int(d/time.Second) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

type callTracker struct {
	lock  sync.Mutex
	calls map[string]*trackedCall
}

// add stores a new call. It returns false if the call is already known (e.g. offer and offer_notice for the same call).
func (ct *callTracker) add(call *trackedCall) bool {
	ct.lock.Lock()
	defer ct.lock.Unlock()
	if ct.calls == nil {
		ct.calls = make(map[string]*trackedCall)
	}
	for id, existing := range ct.calls {
		if time.Since(existing.OfferedAt) > callTrackerMaxAge {
			delete(ct.calls, id)
		}
	}
	if _, exists := ct.calls[call.ID]; exists {
		return false
	}
	ct.calls[call.ID] = call
	return true
}

// transition applies fn to the call with the given ID and returns a snapshot of the call if the state changed.
// Calls that reached a final state are removed from the tracker.
func (ct *callTracker) transition(callID string, fn func(call *trackedCall) bool) *trackedCall {
	ct.lock.Lock()
	defer ct.lock.Unlock()
	call, ok := ct.calls[callID]
	if !ok || !fn(call) {
		return nil
	}
	if call.State != callStateRinging && call.State != callStateAccepted {
		delete(ct.calls, callID)
	}
	snapshot := *call
	return &snapshot
}

// ringingInPortal returns snapshots of all calls in the given portal that haven't been answered yet.
func (ct *callTracker) ringingInPortal(portal networkid.PortalKey) []*trackedCall {
	ct.lock.Lock()
	defer ct.lock.Unlock()
	var out []*trackedCall
	for _, call := range ct.calls {
		if call.Portal == portal && call.State == callStateRinging {
			snapshot := *call
			out = append(out, &snapshot)
		}
	}
	return out
}

func (wa *WhatsAppClient) handleWACallOffer(ctx context.Context, evt *events.CallOffer) bool {
	video := false
	if evt.Data != nil {
		for _, child := range evt.Data.GetChildren() {
			if child.Tag == "video" {
				video = true
				break
			}
		}
	}
	return wa.handleWACallStart(ctx, evt.BasicCallMeta, video, false)
}

func (wa *WhatsAppClient) handleWACallOfferNotice(ctx context.Context, evt *events.CallOfferNotice) bool {
	return wa.handleWACallStart(ctx, evt.BasicCallMeta, evt.Media == "video", evt.Type == "group" || !evt.GroupJID.IsEmpty())
}

func (wa *WhatsAppClient) handleWACallStart(ctx context.Context, meta types.BasicCallMeta, video, group bool) bool {
	if !wa.Main.Config.CallStartNotices || time.Since(meta.Timestamp) > callEventMaxAge {
		return true
	}
	sender, senderAlt := meta.CallCreator, meta.CallCreatorAlt
	if sender.Server == types.DefaultUserServer && senderAlt.IsEmpty() {
		senderAlt, _ = wa.GetStore().LIDs.GetLIDForPN(ctx, sender)
	}
	if sender.Server == types.DefaultUserServer && senderAlt.Server == types.HiddenUserServer {
		wa.UserLogin.Log.Debug().
			Stringer("lid", senderAlt).
			Stringer("pn", sender).
			Str("call_id", meta.CallID).
			Msg("Forced phone number caller to LID in incoming call")
		sender = senderAlt
	}
	chat := meta.GroupJID
	if chat.IsEmpty() {
		chat = sender
	}
	call := &trackedCall{
		ID:        meta.CallID,
		Chat:      chat,
		Portal:    wa.makeWAPortalKey(chat),
		Creator:   sender,
		RejectTo:  meta.CallCreator,
		Video:     video,
		Group:     group || !meta.GroupJID.IsEmpty(),
		State:     callStateRinging,
		OfferedAt: meta.Timestamp,
		MessageID: waid.MakeFakeMessageID(chat, sender, "call-"+meta.CallID),
	}
	if !wa.calls.add(call) {
		return true
	}
	snapshot := *call
	return wa.UserLogin.QueueRemoteEvent(&simplevent.Message[*trackedCall]{
		EventMeta: simplevent.EventMeta{
			Type:         bridgev2.RemoteEventMessage,
			PortalKey:    call.Portal,
			Sender:       wa.makeEventSender(ctx, sender),
			CreatePortal: true,
			Timestamp:    meta.Timestamp,
			StreamOrder:  meta.Timestamp.Unix(),
		},
		Data:               &snapshot,
		ID:                 call.MessageID,
		ConvertMessageFunc: convertCallStart,
	}).Success
}

func convertCallStart(_ context.Context, _ *bridgev2.Portal, _ bridgev2.MatrixAPI, call *trackedCall) (*bridgev2.ConvertedMessage, error) {
	return &bridgev2.ConvertedMessage{
		Parts: []*bridgev2.ConvertedMessagePart{{
			Type:    event.EventMessage,
			Content: call.content(),
		}},
	}, nil
}

func convertCallUpdate(_ context.Context, _ *bridgev2.Portal, _ bridgev2.MatrixAPI, existing []*database.Message, call *trackedCall) (*bridgev2.ConvertedEdit, error) {
	if len(existing) == 0 {
		return nil, fmt.Errorf("call notice not found")
	}
	return &bridgev2.ConvertedEdit{
		ModifiedParts: []*bridgev2.ConvertedEditPart{{
			Part:    existing[0],
			Type:    event.EventMessage,
			Content: call.content(),
		}},
	}, nil
}

func (wa *WhatsAppClient) handleWACallAccept(ctx context.Context, meta types.BasicCallMeta) bool {
	return wa.updateCallState(ctx, meta, func(call *trackedCall) bool {
		if call.State != callStateRinging {
			return false
		}
		call.State = callStateAccepted
		call.AcceptedAt = meta.Timestamp
		return true
	})
}

func (wa *WhatsAppClient) handleWACallReject(ctx context.Context, meta types.BasicCallMeta) bool {
	return wa.updateCallState(ctx, meta, func(call *trackedCall) bool {
		if call.State != callStateRinging {
			return false
		}
		call.State = callStateDeclined
		call.EndedAt = meta.Timestamp
		return true
	})
}

func (wa *WhatsAppClient) handleWACallTerminate(ctx context.Context, evt *events.CallTerminate) bool {
	return wa.updateCallState(ctx, evt.BasicCallMeta, func(call *trackedCall) bool {
		switch call.State {
		case callStateRinging:
			call.State = callStateMissed
		case callStateAccepted:
			call.State = callStateEnded
		default:
			return false
		}
		call.EndedAt = evt.Timestamp
		return true
	})
}

func (wa *WhatsAppClient) updateCallState(ctx context.Context, meta types.BasicCallMeta, fn func(call *trackedCall) bool) bool {
	call := wa.calls.transition(meta.CallID, fn)
	if call == nil {
		return true
	}
	return wa.queueCallUpdate(ctx, call)
}

func (wa *WhatsAppClient) queueCallUpdate(ctx context.Context, call *trackedCall) bool {
	ts := call.EndedAt
	if ts.IsZero() {
		ts = call.AcceptedAt
	}
	if ts.IsZero() {
		ts = time.Now()
	}
	return wa.UserLogin.QueueRemoteEvent(&simplevent.Message[*trackedCall]{
		EventMeta: simplevent.EventMeta{
			Type:      bridgev2.RemoteEventEdit,
			PortalKey: call.Portal,
			Sender:    wa.makeEventSender(ctx, call.Creator),
			Timestamp: ts,
		},
		Data:            call,
		ID:              call.MessageID,
		TargetMessage:   call.MessageID,
		ConvertEditFunc: convertCallUpdate,
	}).Success
}

// DeclineCallsInPortal rejects all ringing WhatsApp calls in the given portal and returns how many were declined.
func (wa *WhatsAppClient) DeclineCallsInPortal(ctx context.Context, portal networkid.PortalKey) (int, error) {
	declined := 0
	for _, call := range wa.calls.ringingInPortal(portal) {
		if err := wa.Client.RejectCall(ctx, call.RejectTo, call.ID); err != nil {
			return declined, fmt.Errorf("failed to reject call %s: %w", call.ID, err)
		}
		updated := wa.calls.transition(call.ID, func(c *trackedCall) bool {
			if c.State != callStateRinging {
				return false
			}
			c.State = callStateDeclinedFromMatrix
			c.EndedAt = time.Now()
			return true
		})
		if updated != nil {
			wa.queueCallUpdate(ctx, updated)
		}
		declined++
	}
	return declined, nil
}
