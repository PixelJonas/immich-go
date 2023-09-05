package host_if

import (
	"context"
	"immich-go/fshelper"
)

type HostConnection interface {
	Close() error
	OpenFS() (fshelper.RemoteFS, error)
	Commander
}

type Commander interface {
	CommandContext(context.Context, string, ...string) (Runner, error)
}

type Runner interface {
	Run() error
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}
