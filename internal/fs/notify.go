package fs

import "errors"

// EventKind describes a filesystem mutation.
type EventKind int

const (
	EventCreate EventKind = iota
	EventWrite
	EventRemove
	EventRename
)

// Event is passed to after hooks.
type Event struct {
	Kind EventKind
	Path string
}

// BeforeHook runs before an operation; return ErrCanceled to abort.
type BeforeHook func(kind EventKind, path string) error

// AfterHook runs after a successful operation.
type AfterHook func(ev Event)

// RegisterBeforeHook adds a before hook on the shared FS.
func (f *FS) RegisterBeforeHook(h BeforeHook) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.before = append(f.before, h)
}

// RegisterAfterHook adds an after hook on the shared FS.
func (f *FS) RegisterAfterHook(h AfterHook) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.after = append(f.after, h)
}

func (f *FS) runBefore(kind EventKind, path string) error {
	f.mu.RLock()
	hooks := append([]BeforeHook(nil), f.before...)
	f.mu.RUnlock()
	for _, h := range hooks {
		if err := h(kind, path); err != nil {
			if errors.Is(err, ErrCanceled) {
				return err
			}
			return err
		}
	}
	return nil
}

func (f *FS) runAfter(ev Event) {
	f.mu.RLock()
	hooks := append([]AfterHook(nil), f.after...)
	f.mu.RUnlock()
	for _, h := range hooks {
		h(ev)
	}
}
