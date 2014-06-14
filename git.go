package vcs

import (
	"strings"
)

var (
	gitFormat = "--format=" + beginCommit + "%H%n%h%n%an%n%ae%n%at%n%s%n%b%n" + endMessage
)

type git struct {
}

func (g *git) Cmd() string {
	return "git"
}

func (g *git) Dir() string {
	return ".git"
}

func (g *git) Head() string {
	return "HEAD"
}

func (g *git) Revision(id string) []string {
	return []string{"show", id, gitFormat}
}

func (g *git) History(since string) []string {
	args := []string{"log", "-p"}
	if since != "" {
		args = append(args, since+"..HEAD")
	}
	args = append(args, gitFormat)
	return args
}

func (g *git) Checkout(rev string) []string {
	if rev == "" {
		rev = "master"
	}
	return []string{"checkout", "-f", rev}
}

func (g *git) Clone(src string, dst string) []string {
	return []string{"clone", src, dst}
}

func (g *git) Update() []string {
	// --rebase fixes the local repo if upstream
	// was rebased.
	return []string{"pull", "-f", "--rebase"}
}

func (g *git) ParseRevisions(_ string, data []byte) ([]*Revision, error) {
	commits := strings.Split(string(data), beginCommit)
	revs := make([]*Revision, 0, len(commits))
	for _, v := range commits {
		if v == "" {
			continue
		}
		lines := strings.Split(v, "\n")
		ts, err := timestamp(lines[4])
		if err != nil {
			return nil, err
		}
		var msg []string
		jj := 6
		for ; jj < len(lines); jj++ {
			if lines[jj] == endMessage {
				break
			}
			msg = append(msg, lines[jj])
		}
		revs = append(revs, &Revision{
			Identifier:      lines[0],
			ShortIdentifier: lines[1],
			Subject:         lines[5],
			Message:         strings.TrimSpace(strings.Join(msg, "\n")),
			Author:          lines[2],
			Email:           lines[3],
			Diff:            strings.TrimSpace(strings.Join(lines[jj+1:], "\n")),
			Timestamp:       ts,
		})
	}
	return revs, nil
}

func (g *git) Branches() []string {
	return []string{"show-ref", "--heads"}
}

func (g *git) Tags() []string {
	return []string{"show-ref", "--tags"}
}

func (g *git) ParseBranches(data []byte) ([]*Branch, error) {
	lines, err := splitTwoValuesLines(data)
	if err != nil {
		return nil, err
	}
	branches := make([]*Branch, len(lines))
	for ii, v := range lines {
		branches[ii] = &Branch{
			Name:     strings.TrimPrefix(v.second, "refs/heads/"),
			Revision: v.first,
		}
	}
	return branches, nil
}

func (g *git) ParseTags(data []byte) ([]*Tag, error) {
	lines, err := splitTwoValuesLines(data)
	if err != nil {
		return nil, err
	}
	tags := make([]*Tag, len(lines))
	for ii, v := range lines {
		tags[ii] = &Tag{
			Name:     strings.TrimPrefix(v.second, "refs/tags/"),
			Revision: v.first,
		}
	}
	return tags, nil
}

func init() {
	Register(&git{})
}
