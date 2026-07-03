package health

import "context"

type Dependency interface {
	Name() string
	Check(ctx context.Context) error
	Weight() int
	Critical() bool
}
