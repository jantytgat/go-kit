package semver

import (
	"fmt"
	"regexp"
)

const (
	validMetadata = `^(?P<commit>[0-9a-zA-Z]{8}).(?P<date>[0-9]{8})$`
)

var regexMetadata = regexp.MustCompile(validMetadata)

type Metadata string

func SplitMetadata(m Metadata) (string, string, error) {
	if !regexMetadata.MatchString(string(m)) {
		return "", "", fmt.Errorf("invalid metadata: %s", m)
	}

	match := regexMetadata.FindStringSubmatch(string(m))
	return match[1], match[2], nil
}
