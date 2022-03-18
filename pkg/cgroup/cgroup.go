package cgroup

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

// fs is used to mock filesystem in unit tests
var fs = afero.NewOsFs()

const (
	// PseudoFsPrefix represents cgroup pseudo-filesystem prefix.
	PseudoFsPrefix = "/sys/fs/cgroup/"
	// v2Indicator represents a file name which indicates that cgropu v2 is enabled on host.
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
	return writeFile(path, strconv.Itoa(os.Getpid()), os.O_WRONLY|os.O_TRUNC)
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

	// 2. Ensure that given controllers are enabled on root level
	requiredControllers := map[string]struct{}{}
	for _, name := range controllers {
		requiredControllers[string(name)] = struct{}{}
	}
	if err := ensureRootControllerAreEnabled(requiredControllers); err != nil {
		return err
	}

	// 3. Enable controllers on child level
	ctrls := strings.Builder{}
	for _, name := range controllers {
		// NOTE: new line will be added at the end of the file, but that's ok.
		ctrls.WriteString(fmt.Sprintf("+%s\n", name))
	}

	childCtrlPath := filepath.Join(dir, controllersFileName)
	err = writeFile(childCtrlPath, ctrls.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
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

	for name, val := range encoded {
		fpath := filepath.Join(dir, name)
		err := writeFile(fpath, fmt.Sprintf("%v", val), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
		if err != nil {
			return err
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
	return gpath, fs.Mkdir(gpath, 0755)
}

func ensureRootControllerAreEnabled(requiredControllersIndex map[string]struct{}) error {
	rootCtrlPath := filepath.Join(PseudoFsPrefix, controllersFileName)
	file, err := fs.OpenFile(rootCtrlPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	var outControllers []string
	removeSign := strings.NewReplacer("+", "", "-", "")
	scanner := bufio.NewScanner(file)

	// find controller which should be preserved
	for scanner.Scan() {
		line := scanner.Text()
		gotCtrl := removeSign.Replace(line)
		if _, found := requiredControllersIndex[gotCtrl]; found {
			// might be conflicting
			continue
		}
		outControllers = append(outControllers, gotCtrl)
	}

	for name := range requiredControllersIndex {
		outControllers = append(outControllers, fmt.Sprintf("+%s", name))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	_, err = file.WriteString(strings.Join(outControllers, "\n"))
	if err != nil {
		return err
	}
	return nil
}
