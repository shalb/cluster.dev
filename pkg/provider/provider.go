package provider

// Activity common interface for module and provisioner.
type Activity interface {
	// Apply module with his defined configuration.
	Deploy() error
	// Destroy infrastructure, created by module.
	Destroy() error
	// Some modules checks.
	Check() (bool, error)
	// Path to directory with module files.
	Path() string
	// Clear module tmp files and cache.
	Clear() error
}

type ActivityDesc struct {
	Category string
	Name     string
}
