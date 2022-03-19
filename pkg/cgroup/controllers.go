package cgroup

type Controller string

const (
	MemoryController Controller = "memory"
	CPUController    Controller = "cpu"
	CPUSetController Controller = "cpuset"
	IOController     Controller = "io"
)
