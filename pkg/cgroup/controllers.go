package cgroup

type Controller string

const (
	MemoryController Controller = "memory"
	CPUController    Controller = "cpu"
	IOController     Controller = "io"
)
