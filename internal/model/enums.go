package model

type AccountMode string

const (
	AccountModeDirect AccountMode = "direct"
	AccountModeRelay  AccountMode = "relay"
)

type PairingState string

const (
	PairingStateUnpaired PairingState = "unpaired"
	PairingStatePending  PairingState = "pending"
	PairingStatePaired   PairingState = "paired"
	PairingStateBlocked  PairingState = "blocked"
)

type InboundMessageStatus string

const (
	InboundStatusQueued    InboundMessageStatus = "queued"
	InboundStatusDelivered InboundMessageStatus = "delivered"
	InboundStatusAcked     InboundMessageStatus = "acked"
	InboundStatusExpired   InboundMessageStatus = "expired"
)

type OutboundMessageStatus string

const (
	OutboundStatusPending OutboundMessageStatus = "pending"
	OutboundStatusSent    OutboundMessageStatus = "sent"
	OutboundStatusFailed  OutboundMessageStatus = "failed"
)

type SessionStatus string

const (
	SessionStatusPendingPairing SessionStatus = "pending_pairing"
	SessionStatusPaired         SessionStatus = "paired"
	SessionStatusExpired        SessionStatus = "expired"
	SessionStatusDisconnected   SessionStatus = "disconnected"
)
