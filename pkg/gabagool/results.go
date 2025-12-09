package gabagool

import "errors"

var (
	ErrCancelled = errors.New("operation cancelled by user")
)

type ListAction int

const (
	ListActionSelected ListAction = iota
	ListActionTriggered
)

type DetailAction int

const (
	DetailActionNone DetailAction = iota
	DetailActionTriggered
	DetailActionConfirmed
	DetailActionCancelled
)
