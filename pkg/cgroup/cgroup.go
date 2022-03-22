package cgroup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// fs is used to mock filesystem in unit tests
var fs = afero.NewOsFs()

const (
	// PseudoFsPrefix represents cgroup pseudo-filesystem prefix.
	PseudoFsPrefix = "/sys/fs/cgroup/"
	// v2Indicator represents a file name which indicates that cgroup v2 is enabled on host.
	v2Indicator = "/sys/fs/cgroup/cgroup.controllers"
	// Procs represents a file name which has the PIDs of all processes belonging to the cgroup. One per line.
	procsFileName = "cgroup.procs"
	// controllers represents a file name which specifies enabled/disabled controllers for a given child group.
	controllersFileName = "cgroup.subtree_control"
)

// IsCgroupV2Enabled returns true if cgroup v2 is enabled.
func IsCgroupV2Enabled() (bool, error) {
	_, err := fs.Stat(v2Indicator)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	default:
		return false, err
	}
}

// AttachCurrentProc attaches current process PID to a given cgroup.
func AttachCurrentProc(cgroup string) error {
	if err := ValidateGroupPath(cgroup); err != nil {
		return err
	}
	path := filepath.Join(cgroup, procsFileName)
	return writeFile(path, fmt.Sprintf("%d\n", os.Getpid()), os.O_WRONLY|os.O_TRUNC)
}

// ValidateGroupPath verifies the cgroup path format.
func ValidateGroupPath(in string) error {
	if !strings.HasPrefix(in, PseudoFsPrefix) {
		return fmt.Errorf("invalid format, cgroup path should start with %q", PseudoFsPrefix)
	}
	return nil
}

// BootstrapParent bootstraps a parent group where a proper controllers are enabled for children.
// TODO(simplification): it works only 1 level deep - /sys/fs/cgroup/{PARENT}/
func BootstrapParent(groupPath string, controllers ...Controller) error {
	// 1. Create parent directory
	dir, err := createCgroupDir(groupPath)
	if err != nil {
		return err
	}

	if len(controllers) == 0 {
		return nil
	}

	ctrls := strings.Builder{}
	for _, name := range controllers {
		ctrls.WriteString(name.Enable())
	}
	ctrlsToEnable := ctrls.String()

	// 2. Ensure that given controllers are enabled on root level
	rootCtrlPath := filepath.Join(PseudoFsPrefix, controllersFileName)
	err = writeFile(rootCtrlPath, ctrlsToEnable, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return err
	}

	// 3. Enable controllers on child level
	childCtrlPath := filepath.Join(dir, controllersFileName)
	err = writeFile(childCtrlPath, ctrlsToEnable, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}

	return nil
}

// BootstrapChild bootstraps a parent child group with given resources' restrictions.
func BootstrapChild(groupPath string, resources Resources) error {
	dir, err := createCgroupDir(groupPath)
	if err != nil {
		return err
	}

	encoded, err := MapResourceToFiles(resources)
	if err != nil {
		return err
	}

	type MultiEntry interface {
		Items() []string
	}

	for name, val := range encoded {
		fpath := filepath.Join(dir, name)

		switch v := val.(type) {
		case MultiEntry:
			for _, item := range v.Items() {
				err := writeFile(fpath, fmt.Sprintf("%v", item), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
				if err != nil {
					return err
				}
			}
		default:
			err := writeFile(fpath, fmt.Sprintf("%v", v), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(path string, data string, flag int) error {
	f, err := fs.OpenFile(path, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(data)
	return err
}

func createCgroupDir(groupPath string) (string, error) {
	gpath := filepath.Clean(groupPath)
	if !strings.HasPrefix(gpath, PseudoFsPrefix) {
		gpath = filepath.Join(PseudoFsPrefix, gpath)
	}

	_, err := os.Stat(gpath)
	switch {
	case err == nil:
		return gpath, nil
	case os.IsNotExist(err):
		return gpath, fs.Mkdir(gpath, 0755)
	default:
		return "", err
	}
}
