package process

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"go.linka.cloud/pm"
	"go.linka.cloud/pm/reexec"
	"go.uber.org/multierr"
	"google.golang.org/grpc"

	"go.linka.cloud/grpc-toolkit/logger"
	"go.linka.cloud/grpc-toolkit/service"
	"go.linka.cloud/grpc-toolkit/signals"
)

var _ pm.Service = (*Child)(nil)

var serviceRegx = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type Childs []*Child

func (c Childs) Register() {
	for _, c := range c {
		c.Register()
	}
}

func (c Childs) WithOpts(opts ...pm.CmdOpt) Childs {
	for _, c := range c {
		c.WithOpts(opts...)
	}
	return c
}

func (c Childs) Close() error {
	var err error
	for _, c := range c {
		err = multierr.Append(err, c.Close())
	}
	return err
}

func NewChild(name string, opts ...service.Option) (*Child, error) {
	if !serviceRegx.MatchString(name) {
		return nil, errors.New("invalid name")
	}
	return &Child{name: name, o: opts}, nil
}

type Child struct {
	name string
	o    []service.Option
	co   []pm.CmdOpt
	c    *pm.Cmd
	m    sync.RWMutex
}

func (c *Child) WithOpts(opts ...pm.CmdOpt) *Child {
	c.co = append(c.co, opts...)
	return c
}

func (c *Child) Serve(ctx context.Context) error {
	c.m.RLock()
	if c.c != nil {
		c.m.RUnlock()
		return pm.ErrAlreadyRunning
	}
	c.m.RUnlock()
	s := c.socket()
	_ = os.Remove(s)
	lis, err := net.ListenUnix("unix", &net.UnixAddr{Name: s})
	if err != nil {
		return err
	}
	defer lis.Close()
	defer os.Remove(s)
	if os.Getenv("PM_NO_FORK") == "1" {
		return c.serve(ctx, lis)
	}
	f, err := lis.File()
	if err != nil {
		return err
	}
	c.m.Lock()
	c.c = pm.ReExec(c.name).WithOpts(pm.WithExtraFiles(f), pm.WithCancel(func(cmd *exec.Cmd) error {
		return cmd.Process.Signal(os.Interrupt)
	}))
	c.m.Unlock()
	defer func() {
		c.m.Lock()
		c.c = nil
		c.m.Unlock()
	}()
	return c.c.Serve(ctx)
}

func (c *Child) Register() {
	reexec.Register(c.name, func() {
		ctx := signals.SetupSignalHandler()
		ctx = logger.Set(ctx, logger.C(ctx).WithField("service", c.String()))
		if err := c.run(ctx); err != nil {
			logger.C(ctx).Fatal(err)
		}
	})
}

func (c *Child) Dial(ctx context.Context, opts ...grpc.DialOption) (grpc.ClientConnInterface, error) {
	conn, err := grpc.DialContext(ctx, "", append(opts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		c.m.RLock()
		defer c.m.RUnlock()
		if c.c == nil {
			return nil, errors.New("not running")
		}
		return net.Dial("unix", c.socket())
	}))...)
	return conn, err
}

func (c *Child) Close() error {
	c.m.Lock()
	defer c.m.Unlock()
	if c.c != nil {
		return c.c.Signal(os.Interrupt)
	}
	return nil
}

func (c *Child) String() string {
	return c.name
}

func (c *Child) socket() string {
	name := strings.NewReplacer("/", "-", ":", "-", " ", "-").Replace(c.name)
	dir := "/tmp"
	if d := os.Getenv("TMPDIR"); d != "" {
		dir = d
	}
	return filepath.Join(dir, name+".sock")
}

func (c *Child) run(ctx context.Context) error {
	f := os.NewFile(3, "conn")
	if f == nil {
		return errors.New("invalid connection file descriptor")
	}
	lis, err := net.FileListener(f)
	if err != nil {
		return err
	}
	defer lis.Close()
	return c.serve(ctx, lis)
}

func (c *Child) serve(ctx context.Context, lis net.Listener) error {
	logger.C(ctx).Infof("starting service")
	s, err := service.New(append(
		c.o,
		service.WithContext(ctx),
		service.WithListener(lis),
	)...)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()
	return s.Start()
	return nil
}
