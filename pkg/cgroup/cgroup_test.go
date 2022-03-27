package cgroup_test

import (
	"os"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mszostok/job-runner/pkg/cgroup"
)

func TestIsCgroupV2Enabled(t *testing.T) {
	t.Run("Should indicate that cgroup v2 is enabled", func(t *testing.T) {
		// given
		tFS := afero.NewMemMapFs()
		revert := cgroup.SetFS(tFS)
		defer revert()

		_, err := tFS.Create("/sys/fs/cgroup/cgroup.controllers")
		require.NoError(t, err)

		// when
		err = cgroup.CheckCgroupV2Enabled()

		// then
		assert.NoError(t, err)
	})

	t.Run("Should indicate that cgroup v2 is disabled", func(t *testing.T) {
		// given
		tFS := afero.NewMemMapFs()
		revert := cgroup.SetFS(tFS)
		defer revert()

		// when
		err := cgroup.CheckCgroupV2Enabled()

		// then
		assert.ErrorIs(t, err, cgroup.ErrCgroupV2NotSupported)
	})

	t.Run("Should indicate that cgroup v2 is disabled", func(t *testing.T) {
		// given
		tFS := &erroneousFS{
			Fs:        afero.NewMemMapFs(),
			StatError: errors.New("some error"),
		}
		revert := cgroup.SetFS(tFS)
		defer revert()

		// when
		err := cgroup.CheckCgroupV2Enabled()

		// then
		assert.ErrorIs(t, err, tFS.StatError)
	})
}

func TestValidateGroupPath(t *testing.T) {
	t.Run("Should return no error for valid path", func(t *testing.T) {
		err := cgroup.ValidateGroupPath("/sys/fs/cgroup/lpr")
		assert.NoError(t, err)
	})

	t.Run("Should return error for invalid path", func(t *testing.T) {
		err := cgroup.ValidateGroupPath("/some/random/path")
		assert.Error(t, err)
	})
}

type erroneousFS struct {
	afero.Fs
	StatError error
}

func (w *erroneousFS) Stat(name string) (os.FileInfo, error) {
	if w.StatError != nil {
		return nil, w.StatError
	}
	return w.Fs.Stat(name)
}
