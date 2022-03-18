package job

import (
	"github.com/mszostok/job-runner/pkg/cgroup"
)

var DefaultProcResources = cgroup.Resources{
	IO: &cgroup.IO{
		Max: []cgroup.IOMax{
			{
				Type:  cgroup.ReadIOPS,
				Major: 8,       // device major
				Minor: 0,       // device minor
				Rate:  1048576, // 1MB (1024^2)
			},
		},
	},
	CPU: &cgroup.CPU{
		// "The first value is the allowed time quota in microseconds for which
		//  all processes collectively in a child group can run during one period.
		//  The second value specifies the length of the period."
		// Translates to: run on the CPU for only 0.1 second of every 1 second.
		Max:  "100000 1000000",
		Cpus: "1",
	},
	Memory: &cgroup.Memory{
		// Memory usage hard limit. If a cgroup's memory usage reaches this limit and
		//	can't be reduced, the OOM killer is invoked in the cgroup.
		// source: https://www.kernel.org/doc/Documentation/cgroup-v2.txt
		Max: 104857600, // 100MB (1024^2*100)
	},
}
