package host

import (
	"context"
	"fmt"
	"immich-go/host/host_if"
	"immich-go/host/local"
	"immich-go/host/ssh"
	"net/url"
)

func Open(ctx context.Context, host string) (host_if.HostConnection, error) {
	u, _ := url.Parse(host)

	switch u.Scheme {
	case "":
		return local.New(ctx, u)
	case "ssh":
		return ssh.New(ctx, u)
	}
	return nil, fmt.Errorf("can't recognize the url:%q", host)
}
