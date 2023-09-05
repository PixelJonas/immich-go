package ssh

import (
	"context"
	"fmt"
	"immich-go/fshelper"
	"immich-go/host/host_if"
	"io/fs"
	"net/url"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
)

type SshConn struct {
	*goph.Client
	ctx context.Context
	u   *url.URL
}

func New(ctx context.Context, u *url.URL) (*SshConn, error) {
	if u.Scheme != "ssh" {
		return nil, fmt.Errorf("unsupported protocol %s: %s", u.Scheme, u.String())
	}
	if u.Host == "" {
		return nil, fmt.Errorf("missing host: %s", u.String())
	}

	ss := strings.SplitN(u.Host, ":", 2)
	c := SshConn{
		ctx: ctx,
		u:   u,
	}
	callback, err := goph.DefaultKnownHosts()
	if err != nil {
		return nil, err
	}

	conf := goph.Config{
		Addr:     ss[0],
		User:     u.User.Username(),
		Port:     22,
		Callback: callback,
	}

	if len(ss) > 1 {
		port, err := strconv.ParseUint(ss[1], 10, 32)
		if err != nil {
			return nil, err
		}
		conf.Port = uint(port)
	}

	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	if conf.User == "" {
		conf.User = user.Username
	}

	pass, set := u.User.Password()
	if set {
		conf.Auth = goph.Password(pass)
	} else {
		keyFile := filepath.Join(user.HomeDir, ".ssh", "id_rsa")
		conf.Auth, err = goph.Key(keyFile, "")
		if err != nil {
			return nil, err
		}
	}

	c.Client, err = goph.NewConn(&conf)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c SshConn) OpenFS() (fshelper.RemoteFS, error) {
	a := sftpAdapter{}
	var err error
	a.Client, err = c.Client.NewSftp()
	return a, err
}

func (c SshConn) CommandContext(ctx context.Context, name string, arg ...string) (host_if.Runner, error) {
	return c.Client.CommandContext(ctx, name, arg...)
}

type sftpAdapter struct {
	*sftp.Client
}

func (a sftpAdapter) Open(name string) (fs.File, error) {
	return a.Client.Open(name)
}

func (a sftpAdapter) Create(name string) (fshelper.Writer, error) {
	return a.Client.Create(name)
}

func (a sftpAdapter) ReadDir(p string) ([]fs.DirEntry, error) {
	fes, err := a.Client.ReadDir(p)
	if err != nil {
		return nil, err
	}
	r := make([]fs.DirEntry, len(fes))
	for i := range fes {
		r[i] = dirEntryAdepter{FileInfo: fes[i]}
	}
	return r, nil
}

type dirEntryAdepter struct {
	fs.FileInfo
}

func (s dirEntryAdepter) Type() fs.FileMode {
	return s.FileInfo.Mode()
}
func (s dirEntryAdepter) Info() (fs.FileInfo, error) {
	return s.FileInfo, nil
}

/*

func (p *sshProxy) docker(ctx context.Context, args ...string) (cmdAdaptor, error) {
	cmd, err := p.sshClient.CommandContext(ctx, "docker", args...)
	return &sshCmd{Cmd: cmd}, err
}

// sshCmd shim
type sshCmd struct {
	*goph.Cmd
}

func (c *sshCmd) StdoutPipe() (io.ReadCloser, error) {
	r, err := c.Cmd.StdoutPipe()
	return io.NopCloser(r), err
}

*/
