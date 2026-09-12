package app

// InitApplicationServices registers what the application layer contributes to the container.
//
// It is empty by design: every application service this module has is built by its resource onion
// in dynamicengines/, because each one needs the composable default the onion constructs. There is
// nothing left for this to register, and the function exists so the module's Init() reads like
// every other module's.
func InitApplicationServices() error {
	return nil
}
