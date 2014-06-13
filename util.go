package vcs

import (
	"net/mail"
	"strconv"
	"strings"
	"time"
)

func timestamp(ts string) (time.Time, error) {
	if dot := strings.IndexByte(ts, '.'); dot >= 0 {
		ts = ts[:dot]
	}
	val, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0).UTC(), nil
}

func splitMessage(lines []string) (string, string) {
	if len(lines) > 1 {
		return strings.TrimSpace(lines[0]), strings.TrimSpace(strings.Join(lines[1:], "\n"))
	}
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), ""
	}
	return "", ""
}

func splitAuthor(s string) (string, string) {
	if addr, err := mail.ParseAddress(s); err == nil {
		return addr.Name, addr.Address
	}
	return s, ""
}

type twoValuesLine struct {
	first  string
	second string
}

func splitTwoValuesLines(data []byte) ([]*twoValuesLine, error) {
	var values []*twoValuesLine
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		line = strings.TrimSpace(line)
		for ii := 0; ii < len(line); ii++ {
			if line[ii] == ' ' {
				values = append(values, &twoValuesLine{line[:ii], strings.TrimSpace(line[ii:])})
				break
			}
		}
	}
	return values, nil
}
