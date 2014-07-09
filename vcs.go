package vcs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	beginCommit = "\uffff"
	endMessage  = "\ufffe"
)

var (
	// ErrNoVCS is returned when no VCS can be found at the given
	// path.
	ErrNoVCS = errors.New("could not find an VCS at the given path")
)

type Revision struct {
	Identifier      string
	ShortIdentifier string
	Subject         string
	Message         string
	Author          string
	Email           string
	Diff            string
	Timestamp       time.Time
}

type Branch struct {
	Name     string
	Revision string
}

type Tag struct {
	Name     string
	Revision string
}

// VCS represents a VCS system on a given
// directory.
type VCS struct {
	// Dir is the absolute path of the repository root.
	Dir   string
	path  string
	iface Interface
}

func (v *VCS) cmd(args []string) ([]byte, error) {
	return v.dirCmd(v.Dir, args)
}

func (v *VCS) dirCmd(dir string, args []string) ([]byte, error) {
	var err error
	if v.path == "" {
		v.path, err = exec.LookPath(v.iface.Cmd())
		if err != nil {
			return nil, err
		}
	}
	cmd := exec.Command(v.path, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// git show-ref will return exit status 1 when
		// there are not tags. Workaround this issue.
		if len(args) > 0 && args[0] == "show-ref" && stdout.Len() == 0 && stderr.Len() == 0 {
			return stdout.Bytes(), nil
		}
		if stderr.Len() > 0 {
			err = errors.New(stderr.String())
		}
		return nil, fmt.Errorf("command %s %s (at dir %s) failed with error: %s", v.path, args, dir, err)
	}
	return stdout.Bytes(), nil
}

// Last returns the last revision from the default branch, known
// as HEAD in git or tip in mercurial.
func (v *VCS) Last() (*Revision, error) {
	return v.Revision(v.iface.Head())
}

// Revision returns the revision identified by the given id, which
// might be either a short or a long revision identifier.
func (v *VCS) Revision(id string) (*Revision, error) {
	data, err := v.cmd(v.iface.Revision(id))
	if err != nil {
		return nil, err
	}
	revs, err := v.iface.ParseRevisions("", data)
	if err != nil {
		return nil, err
	}
	return revs[0], nil
}

// History returns all the revisions from the VCS which are
// newer than the revision identified by the since parameter,
// which might be either a short or a long revision identifier.
// If since is empty, all the history is returned.
func (v *VCS) History(since string) ([]*Revision, error) {
	data, err := v.cmd(v.iface.History(since))
	if err != nil {
		return nil, err
	}
	return v.iface.ParseRevisions(since, data)
}

// Checkout discards all local changes in VCS and updates
// the working copy to the given revision. If no revision
// is given, the one returned by v.Last() will be checked
// out.
func (v *VCS) Checkout(rev string) error {
	_, err := v.cmd(v.iface.Checkout(rev))
	return err
}

// CheckoutAt works like Checkout, but creates a copy of the
// repository at the given directory before perforing the
// Checkout. The returned VCS is the new one created at dir.
func (v *VCS) CheckoutAt(rev string, dir string) (*VCS, error) {
	// Create parent directory.
	p, _ := filepath.Split(dir)
	if err := os.MkdirAll(p, 0755); err != nil {
		return nil, err
	}
	_, err := v.dirCmd("", v.iface.Clone(v.Dir, dir))
	if err != nil {
		return nil, err
	}
	vc, err := NewAt(dir)
	if err != nil {
		return nil, err
	}
	if err := vc.Checkout(rev); err != nil {
		return nil, err
	}
	return vc, err
}

// Update updates the VCS from its upstream. If there's no
// upstream, an error will be returned.
func (v *VCS) Update() error {
	_, err := v.cmd(v.iface.Update())
	return err
}

// Name returns the name of the underlyng VCS interface (e.g.
// git, mercurial, ...).
func (v *VCS) Name() string {
	return v.iface.Cmd()
}

// Branches returns the available branches in the VCS.
func (v *VCS) Branches() ([]*Branch, error) {
	data, err := v.cmd(v.iface.Branches())
	if err != nil {
		return nil, err
	}
	return v.iface.ParseBranches(data)
}

// Branches returns the available tags in the VCS.
func (v *VCS) Tags() ([]*Tag, error) {
	data, err := v.cmd(v.iface.Tags())
	if err != nil {
		return nil, err
	}
	return v.iface.ParseTags(data)
}

// New starts at the given directory and walks up
// until it finds an VCS. If no VCS could be found,
// an error is returned.
func New(dir string) (*VCS, error) {
	cur := dir
	for {
		abs, err := filepath.Abs(cur)
		if err != nil {
			return nil, err
		}
		for _, v := range interfaces {
			if tester, ok := v.(Tester); ok {
				if tester.Test(abs) {
					return &VCS{Dir: abs, iface: v}, nil
				}
				continue
			}
			d := filepath.Join(abs, v.Dir())
			if st, err := os.Stat(d); err == nil && st.IsDir() {
				return &VCS{Dir: abs, iface: v}, nil
			}
		}
		if p := filepath.Dir(cur); p != "" && p != cur {
			cur = p
			continue
		}
		break
	}
	return nil, ErrNoVCS
}

// NewAt works like New, but does not walk up into the the parent
// directories. Id est, it will only succeed if the given directory
// is the root directory of a VCS checkout.
func NewAt(dir string) (*VCS, error) {
	s, err := New(dir)
	if err == nil {
		if s.Dir != dir {
			s = nil
			err = ErrNoVCS
		}
	}
	return s, err
}
