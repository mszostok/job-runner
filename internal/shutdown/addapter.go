package shutdown

// The Func type is an adapter to allow the use of ordinary functions as ShutdownableService.
type Func func()

// Shutdown calls wrapped f().
func (f Func) Shutdown() error {
	f()
	return nil
}
