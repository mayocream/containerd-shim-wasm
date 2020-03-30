// +build linux

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package wasmtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/containerd/cgroups"
	"github.com/containerd/console"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	proc "github.com/containerd/containerd/pkg/process"
	"github.com/containerd/containerd/pkg/stdio"
	"github.com/containerd/containerd/runtime/v2/task"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Exit struct {
	Pid    int
	Status int
}

// NewContainer returns a new runc container
func NewContainer(ctx context.Context, platform stdio.Platform, r *task.CreateTaskRequest, ec chan<- Exit) (c *Container, err error) {
	//ns, err := namespaces.NamespaceRequired(ctx)
	//if err != nil {
	//	return nil, errors.Wrap(err, "create namespace")
	//}

	//var opts options.Options
	//if r.Options != nil {
	//	v, err := typeurl.UnmarshalAny(r.Options)
	//	if err != nil {
	//		return nil, err
	//	}
	//	// TODO: Use custom options type
	//	opts = *v.(*options.Options)
	//}

	//if err := WriteRuntime(r.Bundle, opts.BinaryName); err != nil {
	//	return nil, err
	//}

	b, err := ioutil.ReadFile(filepath.Join(r.Bundle, "config.json"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read spec")
	}

	var spec specs.Spec
	if err := json.Unmarshal(b, &spec); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal spec")
	}

	if spec.Process == nil {
		return nil, errors.Wrapf(errdefs.ErrInvalidArgument, "no process specification")
	}

	var rootRemap string

	rootfs := ""
	for _, m := range r.Rootfs {
		if m.Type == "bind" {
			continue
		}
		if rootfs == "" {
			rootfs = filepath.Join(r.Bundle, "rootfs")
			if err := os.Mkdir(rootfs, 0711); err != nil {
				return nil, err
			}
		}
	}

	defer func() {
		if err != nil && rootfs != "" {
			if err2 := mount.UnmountAll(rootfs, 0); err2 != nil {
				logrus.WithError(err2).Warn("failed to cleanup rootfs mount")
			}
		}
	}()

	if rootfs != "" {
		for _, rm := range r.Rootfs {
			m := &mount.Mount{
				Type:    rm.Type,
				Source:  rm.Source,
				Options: rm.Options,
			}
			if err := m.Mount(rootfs); err != nil {
				return nil, errors.Wrapf(err, "failed to mount rootfs component %v", m)
			}
		}
	} else if len(r.Rootfs) > 0 {
		rootfs = r.Rootfs[0].Source
	} else if spec.Root != nil && spec.Root.Path != "" {
		rootfs = spec.Root.Path
	} else {
		return nil, errors.Wrapf(errdefs.ErrInvalidArgument, "no root provided")
	}
	rootRemap = fmt.Sprintf("/:%s", rootfs)
	if len(spec.Process.Args) > 0 {
		// TODO: bound this
		spec.Process.Args[0] = filepath.Join(rootfs, spec.Process.Args[0])
	}

	p := &process{
		id: r.ID,
		stdio: stdio.Stdio{
			Stdin:    r.Stdin,
			Stdout:   r.Stdout,
			Stderr:   r.Stderr,
			Terminal: r.Terminal,
		},
		remaps: []string{
			rootRemap,
		},
		exited: make(chan struct{}),
		ec:     ec,
		env:    spec.Process.Env,
		args:   spec.Process.Args,
	}

	container := &Container{
		ID:        r.ID,
		Bundle:    r.Bundle,
		process:   p,
		processes: make(map[string]proc.Process),
	}

	pid := p.Pid()
	if pid > 0 {
		cg, err := cgroups.Load(cgroups.V1, cgroups.PidPath(pid))
		if err != nil {
			logrus.WithError(err).Errorf("loading cgroup for %d", pid)
		}
		container.cgroup = cg
	}
	logrus.Infof("process created: %#v", p)

	return container, nil
}

//// ReadRuntime reads the runtime information from the path
//func ReadRuntime(path string) (string, error) {
//	data, err := ioutil.ReadFile(filepath.Join(path, "runtime"))
//	if err != nil {
//		return "", err
//	}
//	return string(data), nil
//}
//
//// WriteRuntime writes the runtime information into the path
//func WriteRuntime(path, runtime string) error {
//	return ioutil.WriteFile(filepath.Join(path, "runtime"), []byte(runtime), 0600)
//}

// Container for operating on a runc container and its processes
type Container struct {
	mu sync.Mutex

	// ID of the container
	ID string
	// Bundle path
	Bundle string
	// Root Remap
	RootRemap string

	ec        chan<- Exit
	cgroup    cgroups.Cgroup
	process   proc.Process
	processes map[string]proc.Process
}

// All processes in the container
func (c *Container) All() (o []proc.Process) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, p := range c.processes {
		o = append(o, p)
	}
	if c.process != nil {
		o = append(o, c.process)
	}
	return o
}

// ExecdProcesses added to the container
func (c *Container) ExecdProcesses() (o []proc.Process) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, p := range c.processes {
		o = append(o, p)
	}
	return o
}

// Pid of the main process of a container
func (c *Container) Pid() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.process.Pid()
}

// Cgroup of the container
func (c *Container) Cgroup() cgroups.Cgroup {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cgroup
}

// CgroupSet sets the cgroup to the container
func (c *Container) CgroupSet(cg cgroups.Cgroup) {
	c.mu.Lock()
	c.cgroup = cg
	c.mu.Unlock()
}

// Process returns the process by id
func (c *Container) Process(id string) (proc.Process, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if id == "" {
		if c.process == nil {
			return nil, errors.Wrapf(errdefs.ErrFailedPrecondition, "container must be created")
		}
		return c.process, nil
	}
	p, ok := c.processes[id]
	if !ok {
		return nil, errors.Wrapf(errdefs.ErrNotFound, "process does not exist %s", id)
	}
	return p, nil
}

// ProcessExists returns true if the process by id exists
func (c *Container) ProcessExists(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.processes[id]
	return ok
}

// ProcessAdd adds a new process to the container
func (c *Container) ProcessAdd(process proc.Process) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processes[process.ID()] = process
}

// ProcessRemove removes the process by id from the container
func (c *Container) ProcessRemove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.processes, id)
}

// Start a container process
func (c *Container) Start(ctx context.Context, r *task.StartRequest) (proc.Process, error) {
	logrus.Info("starting")
	p, err := c.Process(r.ExecID)
	if err != nil {
		return nil, err
	}
	logrus.Info("got process %#v", p)
	if err := p.Start(ctx); err != nil {
		return nil, err
	}

	logrus.Info("done starting", p)
	if c.Cgroup() == nil && p.Pid() > 0 {
		cg, err := cgroups.Load(cgroups.V1, cgroups.PidPath(p.Pid()))
		if err != nil {
			logrus.WithError(err).Errorf("loading cgroup for %d", p.Pid())
		}
		c.cgroup = cg
	}
	logrus.Info("returning process", p)
	return p, nil
}

// Delete the container or a process by id
func (c *Container) Delete(ctx context.Context, r *task.DeleteRequest) (proc.Process, error) {
	p, err := c.Process(r.ExecID)
	if err != nil {
		return nil, err
	}
	if err := p.Delete(ctx); err != nil {
		return nil, err
	}
	if r.ExecID != "" {
		c.ProcessRemove(r.ExecID)
	}
	return p, nil
}

// Exec an additional process
func (c *Container) Exec(ctx context.Context, r *task.ExecProcessRequest) (proc.Process, error) {
	return nil, errors.Wrap(errdefs.ErrNotImplemented, "exec not implemented")
}

// Pause the container
func (c *Container) Pause(ctx context.Context) error {
	return errors.Wrap(errdefs.ErrNotImplemented, "pause not implemented")
}

// Resume the container
func (c *Container) Resume(ctx context.Context) error {
	return errors.Wrap(errdefs.ErrNotImplemented, "resume not implemented")
}

// ResizePty of a process
func (c *Container) ResizePty(ctx context.Context, r *task.ResizePtyRequest) error {
	p, err := c.Process(r.ExecID)
	if err != nil {
		return err
	}
	ws := console.WinSize{
		Width:  uint16(r.Width),
		Height: uint16(r.Height),
	}
	return p.Resize(ws)
}

// Kill a process
func (c *Container) Kill(ctx context.Context, r *task.KillRequest) error {
	p, err := c.Process(r.ExecID)
	if err != nil {
		return err
	}
	return p.Kill(ctx, r.Signal, r.All)
}

// CloseIO of a process
func (c *Container) CloseIO(ctx context.Context, r *task.CloseIORequest) error {
	p, err := c.Process(r.ExecID)
	if err != nil {
		return err
	}
	if stdin := p.Stdin(); stdin != nil {
		if err := stdin.Close(); err != nil {
			return errors.Wrap(err, "close stdin")
		}
	}
	return nil
}

// Checkpoint the container
func (c *Container) Checkpoint(ctx context.Context, r *task.CheckpointTaskRequest) error {
	return errors.Wrap(errdefs.ErrNotImplemented, "checkpoint not implemented")
}

// Update the resource information of a running container
func (c *Container) Update(ctx context.Context, r *task.UpdateTaskRequest) error {
	return errors.Wrap(errdefs.ErrNotImplemented, "update not implemented")
}

// HasPid returns true if the container owns a specific pid
func (c *Container) HasPid(pid int) bool {
	if c.Pid() == pid {
		return true
	}
	for _, p := range c.All() {
		if p.Pid() == pid {
			return true
		}
	}
	return false
}
