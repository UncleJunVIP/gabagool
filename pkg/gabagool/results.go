package gabagool

import "errors"

// Common sentinel errors for all components
var (
	// ErrCancelled is returned when the user cancels an operation (e.g., pressing B/back button)
	ErrCancelled = errors.New("operation cancelled by user")
)

// ListAction represents the action taken when exiting the List component
type ListAction int

const (
	// ListActionSelected indicates the user confirmed their selection (A button or Start button)
	ListActionSelected ListAction = iota
	// ListActionTriggered indicates the user pressed the action button (X button)
	ListActionTriggered
)

// DetailAction represents the action taken when exiting the DetailScreen component
type DetailAction int

const (
	// DetailActionNone indicates the user exited without triggering any action
	DetailActionNone DetailAction = iota
	// DetailActionTriggered indicates the action button was pressed
	DetailActionTriggered
	// DetailActionConfirmed indicates the confirm button was pressed
	DetailActionConfirmed
)
