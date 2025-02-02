package hook

import (
	"errors"
	"testing"
)

func TestHookAddHandlerAndAdd(t *testing.T) {
	calls := ""

	h := Hook[*Event]{}

	h.BindFunc(func(e *Event) error { calls += "1"; return e.Next() })
	h.BindFunc(func(e *Event) error { calls += "2"; return e.Next() })
	h3Id := h.BindFunc(func(e *Event) error { calls += "3"; return e.Next() })
	h.Bind(&Handler[*Event]{
		Id:   h3Id, // should replace 3
		Func: func(e *Event) error { calls += "3'"; return e.Next() },
	})
	h.Bind(&Handler[*Event]{
		Func:     func(e *Event) error { calls += "4"; return e.Next() },
		Priority: -2,
	})
	h.Bind(&Handler[*Event]{
		Func:     func(e *Event) error { calls += "5"; return e.Next() },
		Priority: -1,
	})
	h.Bind(&Handler[*Event]{
		Func: func(e *Event) error { calls += "6"; return e.Next() },
	})
	h.Bind(&Handler[*Event]{
		Func: func(e *Event) error { calls += "7"; e.Next(); return errors.New("test") }, // error shouldn't stop the chain
	})

	h.Trigger(
		&Event{},
		func(e *Event) error { calls += "8"; return e.Next() },
		func(e *Event) error { calls += "9"; return nil }, // skip next
		func(e *Event) error { calls += "10"; return e.Next() },
	)

	if total := len(h.handlers); total != 7 {
		t.Fatalf("Expected %d handlers, found %d", 7, total)
	}

	expectedCalls := "45123'6789"

	if calls != expectedCalls {
		t.Fatalf("Expected calls sequence %q, got %q", expectedCalls, calls)
	}
}
