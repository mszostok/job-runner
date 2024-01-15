package repo

import (
	"os/exec"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/cockroachdb/errors"
)

// JobDefinition represent Job entity.
// TODO(simplification): In the future, we could introduce DTO <-> Model <-> DSO approach to have better decoupling.
type JobDefinition struct {
	// TODO(simplification): In proper scenario Name shouldn't be our ID. Instead we should generate and return associated ID in InsertOutput.
	Name   string `valid:"required"`
	Tenant string `valid:"required"`
	// TODO(simplification): In proper scenario Cmd shouldn't be on InsertInput as it's not possible to store it in external backend. We should have only PID.
	Cmd         *exec.Cmd `valid:"required"`
	Status      string    `valid:"required"`
	ExitCode    int
	RunFinished chan struct{}
}

// Repository contains functionality to manipulate Job objects in repository.
type Repository struct {
	// TODO(simplification): this should be an external storage, e.g Redis.
	// There could be also some shim driver so the backend can be easily swappable.
	store map[string]*JobDefinition
	mu    sync.RWMutex

	validate func(in interface{}) error
}

// NewInMemory creates new in memory Repository instance.
func NewInMemory() *Repository {
	return &Repository{
		store: map[string]*JobDefinition{},
		validate: func(in interface{}) error {
			_, err := govalidator.ValidateStruct(in)
			return err
		},
	}
}

type InsertInput struct {
	Job *JobDefinition
}

// Insert inserts a new Cmd to the storage. It is thread safe.
func (r *Repository) Insert(in InsertInput) error {
	if err := r.validate(in); err != nil {
		return errors.Wrap(err, "while validating input")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.store[in.Job.Name]; found {
		return NewIDCConflictError(in.Job.Name)
	}

	r.store[in.Job.Name] = in.Job
	return nil
}

// GetInput contains parameters necessary to execute Get operation on repository.
type GetInput struct {
	Name string `valid:"required"`
}

// GetOutput contains parameters returned from Get operation on repository.
type GetOutput struct {
	Job *JobDefinition
}

// Get returns job from repository that matches given constrains. It is thread safe.
// Returns NotFoundError when object was not found.
func (r *Repository) Get(in GetInput) (GetOutput, error) {
	if err := r.validate(in); err != nil {
		return GetOutput{}, errors.Wrap(err, "while validating input")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	job, found := r.store[in.Name]
	if !found {
		return GetOutput{}, NewNotFoundError(in.Name)
	}

	return GetOutput{
		Job: job,
	}, nil
}

// UpdateInput contains parameters necessary to execute Update operation on repository
type UpdateInput struct {
	Name string `valid:"required"`

	Status   string
	ExitCode int
}

// UpdateOutput contains parameters returned from Update operation on repository
type UpdateOutput struct{}

// Update updates Cmd in repository, returns NotFoundError in case the object is not found - it's not an upsert.
// It is thread safe.
func (r *Repository) Update(in UpdateInput) error {
	if err := r.validate(in); err != nil {
		return errors.Wrap(err, "while validating input")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	old, found := r.store[in.Name]
	if !found {
		return NewNotFoundError(in.Name)
	}
	old.Status = in.Status
	old.ExitCode = in.ExitCode
	r.store[in.Name] = old

	return nil
}

// GetJobTenantInput contains parameters necessary to execute GetJobTenant operation on repository.
type GetJobTenantInput struct {
	Name string `valid:"required"`
}

// GetJobTenantOutput contains parameters returned from GetJobTenant operation on repository.
type GetJobTenantOutput struct {
	Name   string
	Tenant string
}

// GetJobTenant returns Job's tenant, or NotFoundError.
// It is thread safe.
func (r *Repository) GetJobTenant(in GetJobTenantInput) (GetJobTenantOutput, error) {
	if err := r.validate(in); err != nil {
		return GetJobTenantOutput{}, errors.Wrap(err, "while validating input")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	job, found := r.store[in.Name]
	if !found {
		return GetJobTenantOutput{}, NewNotFoundError(in.Name)
	}

	return GetJobTenantOutput{
		Name:   job.Name,
		Tenant: job.Tenant,
	}, nil
}
