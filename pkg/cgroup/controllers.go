package cgroup

import "fmt"

// Controller represents cgroup v2 supported controllers.
type Controller string

const (
	MemoryController Controller = "memory"
	CPUController    Controller = "cpu"
	CPUSetController Controller = "cpuset"
	IOController     Controller = "io"
)

func (c Controller) Enable() string {
	return fmt.Sprintf("+%s ", c)
}

func (c Controller) Disable() string {
	return fmt.Sprintf("-%s ", c)
}
