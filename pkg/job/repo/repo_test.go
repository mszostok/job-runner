package repo_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mszostok/job-runner/pkg/job/repo"
)

// TODO(simplification): test the rest methods too.
func TestInsert_ValidateInput(t *testing.T) {
	t.Run("failures", func(t *testing.T) {
		tests := map[string]struct {
			name   string
			tenant string
			status string
			job    *exec.Cmd

			expErrMsg string
		}{
			"missing Name": {
				tenant: "bar",
				job:    exec.Command("test"),
				status: "xyz",

				expErrMsg: "while validating input: Job.Name: non zero value required",
			},
			"missing Tenant": {
				name:   "foo",
				job:    exec.Command("test"),
				status: "xyz",

				expErrMsg: "while validating input: Job.Tenant: non zero value required",
			},
			"missing Cmd": {
				name:   "foo",
				tenant: "bar",
				status: "xyz",

				expErrMsg: "while validating input: Job.Cmd: non zero value required",
			},

			"missing Status": {
				name:   "foo",
				tenant: "bar",
				job:    exec.Command("test"),

				expErrMsg: "while validating input: Job.Status: non zero value required",
			},
			"missing Name, Tenant, Status, and Cmd": {
				// no fields are set
				expErrMsg: "while validating input: Job.Cmd: non zero value required;Job.Name: non zero value required;Job.Status: non zero value required;Job.Tenant: non zero value required",
			},
		}
		for tn, tc := range tests {
			tc := tc // copy value, to ensure that parallel execution is safe
			t.Run(tn, func(t *testing.T) {
				t.Parallel()

				// given
				svc := repo.NewInMemory()

				job := &repo.JobDefinition{
					Name:   tc.name,
					Tenant: tc.tenant,
					Cmd:    tc.job,
					Status: tc.status,
				}

				// when
				err := svc.Insert(repo.InsertInput{Job: job})

				// then
				assert.EqualError(t, err, tc.expErrMsg)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// given
		svc := repo.NewInMemory()

		// when
		err := svc.Insert(repo.InsertInput{
			Job: &repo.JobDefinition{
				Name:   "foo",
				Tenant: "bar",
				Cmd:    exec.Command("test"),
				Status: "xyz",
			},
		})

		// then
		assert.NoError(t, err)
	})
}

func TestStorageLifecycle(t *testing.T) {
	t.Parallel()
	// given
	svc := repo.NewInMemory()

	job := &repo.JobDefinition{
		Name:   "foo",
		Tenant: "bar",
		Cmd:    exec.Command("test"),
		Status: "xyz",
	}

	// when
	err := svc.Insert(repo.InsertInput{Job: job})
	require.NoError(t, err)

	// then
	out, err := svc.Get(repo.GetInput{Name: job.Name})
	require.NoError(t, err)
	assert.EqualValues(t, job, out.Job)

	// when
	tenantOut, err := svc.GetJobTenant(repo.GetJobTenantInput{Name: job.Name})

	// then
	require.NoError(t, err)
	assert.EqualValues(t, tenantOut.Name, job.Name)
	assert.EqualValues(t, tenantOut.Tenant, job.Tenant)

	// when
	err = svc.Update(repo.UpdateInput{
		Name:     job.Name,
		Status:   "UPDATED",
		ExitCode: 42,
	})
	require.NoError(t, err)

	// then
	expUpdatedJob := *job
	expUpdatedJob.Status = "UPDATED"
	expUpdatedJob.ExitCode = 42
	out, err = svc.Get(repo.GetInput{Name: job.Name})
	require.NoError(t, err)
	require.NotNil(t, out.Job)
	assert.EqualValues(t, expUpdatedJob, *out.Job)
}
