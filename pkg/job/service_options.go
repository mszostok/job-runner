package job

// ServiceOption provides an option to configure Service instance.
type ServiceOption func(cfg *Service)

// WithoutCgroup disables:
// - creating a dedicated cgroup for executed Job,
// - and execution via child process.
func WithoutCgroup() ServiceOption {
	return func(cfg *Service) {
		cfg.createProcCmd = directProcExecution
	}
}
