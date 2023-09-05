package ssh_test

import (
	"context"
	"immich-go/fshelper"
	"immich-go/host/ssh"
	"io/fs"
	"net/url"
	"testing"
	"time"
)

func TestSfpFS(t *testing.T) {
	host, _ := url.Parse("ssh://root@192.168.10.23:22")
	ctx := context.Background()
	conn, err := ssh.New(ctx, host)
	if err != nil {
		t.Error(err)
	}

	sftpConn, err := conn.OpenFS()
	if err != nil {
		t.Error(err)
	}

	ll, err := sftpConn.ReadDir("/etc")
	if err != nil {
		t.Error(err)
	}
	if len(ll) < 100 {
		t.Errorf("ReadDir fails?")
	}

	ts := time.Now().String()
	f, err := sftpConn.Create("/root/canary.txt")
	if err != nil {
		t.Error(err)
	}
	_, err = fshelper.WriteString(f, ts)
	if err != nil {
		t.Error(err)
	}
	f.Close()

	sftpConn.Close()
	conn.Close()
	if err != nil {
		t.Error(err)
	}

	conn, err = ssh.New(ctx, host)
	if err != nil {
		t.Error(err)
	}

	sftpConn, err = conn.OpenFS()
	if err != nil {
		t.Error(err)
	}

	b, err := fs.ReadFile(sftpConn, "/root/canary.txt")
	if err != nil {
		t.Error(err)
	}
	s := string(b)
	if s != ts {
		t.Errorf("expected %s, got %s", ts, s)
	}

	err = sftpConn.Remove("/root/canary.txt")
	if err != nil {
		t.Error(err)
	}
}
