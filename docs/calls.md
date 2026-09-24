# WhatsApp calls over Matrix

Goal: WhatsApp voice and video calls should work natively from Matrix clients
(Element, Element X, FluffyChat, …), without either the caller or the callee
noticing that a bridge is involved.

This document describes what is implemented today, what is missing, and the plan
to get to full audio/video bridging.

## Status

| Feature | Status |
|---|---|
| Incoming 1:1 call → notice in the Matrix room (voice/video detected) | ✅ implemented |
| Incoming group call → notice in the group portal | ✅ implemented |
| Notice is edited live: answered on another device / missed / declined / ended incl. duration | ✅ implemented |
| Decline a ringing WhatsApp call from Matrix (`!wa decline-call`) | ✅ implemented |
| Native ringing in Matrix clients (`m.call.invite` / MatrixRTC ring) | ❌ blocked on media (see below) |
| Answering a WhatsApp call in a Matrix client (audio) | ❌ blocked on media |
| Video | ❌ blocked on media |
| Starting a WhatsApp call from Matrix | ❌ blocked on media |

All implemented parts live in `pkg/connector/calls.go` and are enabled by the
existing `call_start_notices` option in the network config.

## Why audio/video is not bridged yet

The bridge talks to WhatsApp through [whatsmeow], which emulates WhatsApp Web
("linked device"). whatsmeow (as of September 2026) only exposes call
*signalling*: it emits `CallOffer`, `CallAccept`, `CallPreAccept`, `CallTransport`,
`CallRelayLatency`, `CallTerminate` and `CallReject` events and can send a
`reject`. It contains **no VoIP media stack**, so nothing in the Go ecosystem can
currently answer a WhatsApp call and exchange audio or video.

Ringing a Matrix client without being able to carry the media afterwards would be
worse than the notice: the user would pick up and hear silence while the
WhatsApp caller keeps ringing. That is why native ringing is only enabled
together with the media gateway.

[whatsmeow]: https://github.com/tulir/whatsmeow

## Target architecture

```
 WhatsApp phone/app                 mautrix-whatsapp                       Matrix client
 ──────────────────     ┌─────────────────────────────────────┐    ────────────────────────
                        │  whatsmeow   ─ call signalling ─┐   │
  <call> offer/accept ◄─┼─►  (existing, calls.go)         │   │   m.call.invite/answer/
  relay + SRTP media  ◄─┼─►  WA VoIP engine (NEW) ◄──► media ◄─┼─► candidates/hangup (1:1)
                        │                            gateway  │   or MatrixRTC + LiveKit
                        │                     (NEW, pion)     │   (Element Call / Element X)
                        └─────────────────────────────────────┘
```

### 1. WhatsApp VoIP engine (the hard part)

Needs to be written from scratch (ideally upstreamed into whatsmeow) and must
cover, per the observable protocol:

* Building and answering `<call>` stanzas (`offer`, `preaccept`, `accept`,
  `transport`, `relaylatency`, `terminate`) including audio/video capability
  nodes.
* Exchanging the per-call master key end-to-end encrypted with the Signal
  sessions whatsmeow already maintains (`enc` nodes).
* Connecting to the WhatsApp relay servers announced in the offer (relay
  tokens, STUN-like binding / latency probing).
* Deriving SRTP keys from the call key and sending/receiving SRTP.
* Codecs: Opus for audio, and WhatsApp's video codec with its RTP payload format
  (H.264 or VP8 depending on the peer).

None of this is documented publicly. It has to be reverse-engineered against real
clients and will break whenever WhatsApp changes it. **This cannot be built or
verified without real WhatsApp accounts and devices for testing.**

### 2. Media gateway (Matrix side, feasible with existing libraries)

A [pion/webrtc] peer inside the bridge that:

* **1:1 rooms (legacy VoIP, MSC2746):** the ghost user sends `m.call.invite`
  with an SDP offer when WhatsApp rings; `m.call.answer` from the Matrix user →
  WhatsApp `accept`; `m.call.hangup`/`m.call.reject` → WhatsApp
  `terminate`/`reject`. Works with Element Web/Desktop/Android/iOS.
* **MatrixRTC (MSC4143/MSC4195, Element X/Element Call):** the bridge joins the
  room's LiveKit SFU as a participant and sends an `m.rtc.notification` ring.
* Passes Opus through unchanged when possible (WhatsApp and WebRTC both use Opus)
  and only transcodes video when the codecs differ.
* TURN configuration taken from the homeserver (`/voip/turnServer`).

[pion/webrtc]: https://github.com/pion/webrtc

### 3. Outgoing calls from Matrix

Once 1 and 2 exist: `m.call.invite` to a WhatsApp ghost → WhatsApp `offer` to
the contact's devices, with the same media path in reverse.

## Plan

| Phase | Content | Can be verified in CI? |
|---|---|---|
| 0 | Update fork to current upstream, source-based Docker build | ✅ done |
| 1 | Call state sync + decline from Matrix (this change) | ✅ unit tests; live test needs a WhatsApp account |
| 2 | Matrix media gateway (pion) behind a feature flag, tested with a loopback/echo peer | ✅ mostly |
| 3 | WhatsApp VoIP engine: signalling + key exchange + relay + SRTP audio | ❌ needs real devices and reverse engineering |
| 4 | Wire 2 + 3 together: native ringing, answering, audio | ❌ needs real devices |
| 5 | Video, outgoing calls, group calls | ❌ needs real devices |

## Risks

* WhatsApp's terms of service forbid unofficial clients; a VoIP implementation
  that behaves differently from the official apps is easier to detect and may
  get the account banned.
* The VoIP protocol is undocumented and changes without notice.
* Media transcoding (video) costs significant CPU per call.
