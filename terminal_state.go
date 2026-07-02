package cpace

import (
	"fmt"
	"sync"
)

type singleUseCore interface {
	*initiatorCore | *responderCore
	clear()
}

type singleUseState[C singleUseCore] struct {
	mu   sync.Mutex
	used bool

	// core is assigned once at construction and never reassigned or nil'd:
	// clear() zeroes and nils the core's fields, not this pointer. The terminal
	// guard relies on pointer stability so value copies share one terminal state.
	core C

	uninitialized string
}

func newSingleUseState[C singleUseCore](core C, uninitialized string) *singleUseState[C] {
	return &singleUseState[C]{core: core, uninitialized: uninitialized}
}

func (s *singleUseState[C]) claimFinish() (C, error) {
	var zero C
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.used {
		return zero, ErrStateUsed
	}
	if s.core == nil {
		return zero, s.uninitializedError()
	}
	s.used = true
	return s.core, nil
}

func (s *singleUseState[C]) claimClose() (C, error) {
	var zero C
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.used {
		return zero, nil
	}
	if s.core == nil {
		return zero, s.uninitializedError()
	}
	s.used = true
	return s.core, nil
}

// closeCore claims the state for Close and clears the claimed core's
// persistent secrets. Idempotent: a claim lost to an earlier terminal
// operation returns nil with no core to clear.
func (s *singleUseState[C]) closeCore() error {
	core, err := s.claimClose()
	if err != nil {
		return err
	}
	if core == nil {
		return nil
	}
	core.clear()
	return nil
}

func (s *singleUseState[C]) uninitializedError() error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, s.uninitialized)
}
