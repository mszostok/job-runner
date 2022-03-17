package job

import "fmt"

type IOType string

const (
	ReadBPS   IOType = "rbps"
	WriteBPS  IOType = "wbps"
	ReadIOPS  IOType = "riops"
	WriteIOPS IOType = "wiops"
)

type Resources struct {
	IO     IO
	CPU    CPU
	Memory Memory
}

type IO struct {
	Max []Entry
}

type Entry struct {
	Type  IOType
	Major int64
	Minor int64
	Rate  uint64
}

func (e Entry) String() string {
	return fmt.Sprintf("%d:%d %s=%d", e.Major, e.Minor, e.Type, e.Rate)
}

type CPU struct {
	Max  string
	Cpus string
	Mems string
}

type Memory struct {
	Min *int64
	Max *int64
}
