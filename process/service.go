package process

import (
	"context"

	"go.linka.cloud/pm"

	"go.linka.cloud/grpc-toolkit/service"
)

var _ pm.Service = (*Service)(nil)

var (
	Notify = pm.Notify
)

func NewService(name string, opts ...service.Option) *Service {
	return &Service{name: name, o: opts}
}

type Service struct {
	name string
	o    []service.Option
}

func (s *Service) Serve(ctx context.Context) error {
	svc, err := service.New(s.opts(ctx)...)
	if err != nil {
		return err
	}
	defer svc.Close()
	pm.Notify(ctx, pm.StatusStarting)
	defer func() {
		pm.Notify(ctx, pm.StatusStopped)
	}()
	return svc.Start()
}

func (s *Service) String() string {
	return s.name
}

func (s *Service) opts(ctx context.Context) []service.Option {
	return append(s.o,
		service.WithName(s.name),
		service.WithContext(ctx),
		service.WithAfterStart(func() error {
			pm.Notify(ctx, pm.StatusRunning)
			return nil
		}),
		service.WithBeforeStop(func() error {
			pm.Notify(ctx, pm.StatusStopping)
			return nil
		}),
	)
}
