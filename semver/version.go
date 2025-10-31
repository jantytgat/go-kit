package semver

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

const (
	// https://semver.org/ && https://regex101.com/r/Ly7O1x/3/
	validSemVer = `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
)

var regexSemver = regexp.MustCompile(validSemVer)

type Version struct {
	Major      int64
	Minor      int64
	Patch      int64
	PreRelease PreRelease
	Metadata   Metadata
}

func (v Version) Commit() string {
	commit, _, err := SplitMetadata(v.Metadata)
	if err != nil {
		return string(v.Metadata)
	}
	return commit
}

func (v Version) Date() string {
	_, date, err := SplitMetadata(v.Metadata)
	if err != nil {
		return string(v.Metadata)
	}
	return date
}

func (v Version) Release() string {
	switch v.PreRelease {
	case "":
		return fmt.Sprint("stable")
	default:
		return string(v.PreRelease)
	}
}

func (v Version) String() string {
	var buf bytes.Buffer

	_, _ = fmt.Fprintf(&buf, "%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.PreRelease != "" {
		_, _ = fmt.Fprintf(&buf, "-%s", v.PreRelease)
	}

	if v.Metadata != "" {
		_, _ = fmt.Fprintf(&buf, "+%s", v.Metadata)
	}

	return buf.String()
}

func (v Version) Number() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func Parse(v string) (Version, error) {
	if !regexSemver.MatchString(v) {
		return Version{}, oopsBuilder.With("version", v).New("invalid semver version")
	}

	match := regexSemver.FindStringSubmatch(v)
	matchMap := make(map[string]string)
	for i, name := range regexSemver.SubexpNames() {
		if i != 0 && name != "" {
			matchMap[name] = match[i]
		}
	}

	var err error
	var major int64
	var minor int64
	var patch int64
	var preRelease PreRelease
	var metadata Metadata

	if major, err = strconv.ParseInt(matchMap["major"], 10, 64); err != nil {
		return Version{}, oopsBuilder.With("version", v).Wrapf(err, "invalid major number")
	}

	if minor, err = strconv.ParseInt(matchMap["minor"], 10, 64); err != nil {
		return Version{}, oopsBuilder.With("version", v).Wrapf(err, "invalid minor number")
	}

	if patch, err = strconv.ParseInt(matchMap["patch"], 10, 64); err != nil {
		return Version{}, oopsBuilder.With("version", v).Wrapf(err, "invalid patch number")
	}

	if matchMap["prerelease"] != "" {
		preRelease = PreRelease(matchMap["prerelease"])
	}

	if matchMap["buildmetadata"] != "" {
		metadata = Metadata(matchMap["buildmetadata"])
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		Metadata:   metadata,
	}, nil
}
