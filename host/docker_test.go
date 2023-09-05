package host

import (
	"context"
	"immich-go/host/ssh"
	"net/url"

	"testing"
)

func TestDockerPS(t *testing.T) {
	host, _ := url.Parse("ssh://root@192.168.10.23:22")
	ctx := context.Background()
	conn, err := ssh.New(ctx, host)

	if err != nil {
		t.Error(err)
		return
	}
	dc, err := NewDockerConnection(ctx, conn)
	if err != nil {
		t.Error(err)
		return
	}
	ll, err := dc.DockerPS(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	for _, l := range ll {
		t.Log(l)
	}
}

func TestDockerInspect(t *testing.T) {
	host, _ := url.Parse("ssh://root@192.168.10.23:22")
	ctx := context.Background()
	conn, err := ssh.New(ctx, host)

	if err != nil {
		t.Error(err)
		return
	}
	dc, err := NewDockerConnection(ctx, conn)
	if err != nil {
		t.Error(err)
		return
	}
	p, err := dc.ImmichLibrary(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(p)
}
