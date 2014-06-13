package vcs

import (
	"strings"
)

var (
	hgFormat = "--template=" + beginCommit + "{node}\n{node|short}\n{author}\n{date}\n{desc}\n" + endMessage + "\n"
)

type mercurial struct {
}

func (m *mercurial) Cmd() string {
	return "hg"
}

func (m *mercurial) Dir() string {
	return ".hg"
}

func (m *mercurial) Head() string {
	return "tip"
}

func (m *mercurial) Revision(id string) []string {
	return []string{"log", "-pr", id, hgFormat}
}

func (m *mercurial) History(since string) []string {
	args := []string{"log", "-p", "-b", "default"}
	if since != "" {
		args = append(args, "-r")
		args = append(args, "tip:"+since)
	}
	args = append(args, hgFormat)
	return args
}

func (m *mercurial) Checkout(rev string) []string {
	return []string{"update", "-C", "-r", rev}
}

func (m *mercurial) Clone(src string, dst string) []string {
	return []string{"clone", src, dst}
}

func (m *mercurial) Update() []string {
	return []string{"pull", "-u"}
}

func (m *mercurial) ParseRevisions(since string, data []byte) ([]*Revision, error) {
	commits := strings.Split(string(data), beginCommit)
	revs := make([]*Revision, 0, len(commits))
	for _, v := range commits {
		if v == "" {
			continue
		}
		lines := strings.Split(v, "\n")
		author, email := splitAuthor(lines[2])
		ts, err := timestamp(lines[3])
		if err != nil {
			return nil, err
		}
		var msg []string
		jj := 4
		for ; jj < len(lines); jj++ {
			if lines[jj] == endMessage {
				break
			}
			msg = append(msg, lines[jj])
		}
		subject, message := splitMessage(msg)
		revs = append(revs, &Revision{
			Identifier:      lines[0],
			ShortIdentifier: lines[1],
			Subject:         subject,
			Message:         message,
			Author:          author,
			Email:           email,
			Diff:            strings.TrimSpace(strings.Join(lines[jj+1:], "\n")),
			Timestamp:       ts,
		})
	}
	// Mercurial includes both ends, must remove last
	// commit if since != ""
	if since != "" && len(revs) > 0 {
		revs = revs[:len(revs)-1]
	}
	return revs, nil
}

func (m *mercurial) Branches() []string {
	return []string{"branches"}
}

func (m *mercurial) Tags() []string {
	return []string{"tags"}
}

func (m *mercurial) ParseBranches(data []byte) ([]*Branch, error) {
	lines, err := splitTwoValuesLines(data)
	if err != nil {
		return nil, err
	}
	branches := make([]*Branch, len(lines))
	for ii, v := range lines {
		branches[ii] = &Branch{
			Name:     v.first,
			Revision: v.second,
		}
	}
	return branches, nil
}

func (m *mercurial) ParseTags(data []byte) ([]*Tag, error) {
	lines, err := splitTwoValuesLines(data)
	if err != nil {
		return nil, err
	}
	tags := make([]*Tag, len(lines))
	for ii, v := range lines {
		tags[ii] = &Tag{
			Name:     v.first,
			Revision: v.second,
		}
	}
	return tags, nil
}

func init() {
	Register(&mercurial{})
}
