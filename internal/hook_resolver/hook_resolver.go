package hook_resolver

// Resolver defines a common interface for a Hook event (see [Event]).
type Resolver interface {
	// Next triggers the next handler in the hook's chain (if any).
	Next() error

	NextFunc() func() error

	// PreventDefault prevents the native handler from being called.
	PreventDefault()

	SetNextFunc(f func() error)
}

var _ Resolver = (*Event)(nil)

// Event implements [Resolver] and it is intended to be used as a base
// Hook event that you can embed in your custom typed event structs.
//
// Example:
//
//	type CustomEvent struct {
//		hook.Event
//
//		SomeField int
//	}
type Event struct {
	next           func() error
	preventDefault func()

	DefaultPrevented bool `json:"defaultPrevented"`
}

// Next calls the next hook handler.
func (e *Event) Next() error {
	if e.next != nil {
		return e.next()
	}
	return nil
}

func (e *Event) PreventDefault() {
	e.DefaultPrevented = true
}

// NextFunc returns the function that Next calls.
func (e *Event) NextFunc() func() error {
	return e.next
}

// SetNextFunc sets the function that Next calls.
func (e *Event) SetNextFunc(f func() error) {
	e.next = f
}
