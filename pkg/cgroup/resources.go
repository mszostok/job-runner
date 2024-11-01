package cgroup

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type IOType string

const (
	// ReadBPS represents read bytes per second.
	ReadBPS IOType = "rbps"
	// WriteBPS represents write bytes per second.
	WriteBPS IOType = "wbps"
	// ReadIOPS represents read IO operations per second.
	ReadIOPS IOType = "riops"
	// WriteIOPS represents write IO operations per second.
	WriteIOPS IOType = "wiops"
)

type (
	// Resources represents supported resource configuration.
	Resources struct {
		IO     *IO
		CPU    *CPU
		Memory *Memory
	}

	// CPU represents settings for CPU controller.
	CPU struct {
		Max  string `filename:"cpu.max,omitempty"`
		Cpus string `filename:"cpuset.cpus,omitempty"`
		Mems string `filename:"cpuset.mems,omitempty"`
	}

	// Memory represents settings for Memory controller.
	Memory struct {
		// Takes a memory size in bytes
		Min int64 `filename:"memory.min,omitempty"`
		Max int64 `filename:"memory.max,omitempty"`
	}

	// IO represents settings for IO controller.
	IO struct {
		Max IOMax `filename:"io.max,omitempty"`
	}

	// IOMax represents IOMax collection.
	IOMax []IOMaxEntry

	// IOMaxEntry represents settings for IO limits.
	IOMaxEntry struct {
		Type  IOType
		Major int64
		Minor int64
		Rate  uint64
	}
)

// String returns a data in io.max supported format.
func (e IOMaxEntry) String() string {
	return fmt.Sprintf(`%d:%d %s=%d`, e.Major, e.Minor, e.Type, e.Rate)
}

// Items returns list of IO limits. It doesn't remove duplicated entries (same IOType).
func (io IOMax) Items() []string {
	var out []string
	for _, item := range io {
		out = append(out, item.String())
	}
	return out
}

// MapResourceToFiles returns resources' settings indexed by a proper filename.
//
// NOTE: Experimental approach - this function is "generic" but at the same time requires 3rd party lib (which uses reflection).
// This can be easily changed, and we can attach a dedicated method to each resource entity,
// similar to: https://github.com/containerd/cgroups/blob/2e502f6b9e43588a1108ebdd04c51ad2b04050f0/v2/cpu.go#L57
//
// But I wanted to try this approach as it adds an easy way to associate a given field with a proper file which IMO give more friendly API:
//   Min int64 `filename:"memory.min,omitempty"`
//
func MapResourceToFiles(resources Resources) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	config := &mapstructure.DecoderConfig{
		Result:  &out,
		TagName: "filename",
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}

	// Inspired by: https://go.dev/blog/errors-are-values
	var decodeErr error
	decode := func(in interface{}) {
		if in == nil || decodeErr != nil {
			return
		}
		decodeErr = decoder.Decode(in)
	}

	decode(resources.IO)
	decode(resources.CPU)
	decode(resources.Memory)

	if decodeErr != nil {
		return nil, decodeErr
	}
	return out, nil
}
