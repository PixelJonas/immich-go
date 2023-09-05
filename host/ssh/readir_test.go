package ssh_test

import (
	"bytes"
	"context"
	"immich-go/host/ssh"
	"io"
	"net/url"
	"testing"
)

func TestSshReadDir(t *testing.T) {
	u, _ := url.Parse("ssh://root@192.168.10.23")
	ctx := context.Background()
	c, err := ssh.New(ctx, u)
	if err != nil {
		t.Error(err)
		return
	}

	fsys, err := c.OpenFS()
	if err != nil {
		t.Error(err)
		return
	}

	_, err = fsys.ReadDir("/etc")
	if err != nil {
		t.Error(err)
		return
	}

	f, err := fsys.Open("/etc/passwd")
	if err != nil {
		t.Error(err)
		return
	}
	b := bytes.NewBuffer(nil)

	n, err := io.Copy(b, f)
	t.Logf("read %d, err:%s", n, err)
	f.Close()
}
