package cgroup

// Controller represents cgroup v2 supported controllers.
type Controller string

const (
	MemoryController Controller = "memory"
	CPUController    Controller = "cpu"
	CPUSetController Controller = "cpuset"
	IOController     Controller = "io"
)
