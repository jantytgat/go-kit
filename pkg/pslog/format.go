package pslog

import (
	"fmt"
	"strings"
)

const (
	TextLogFormat Format = iota
	JsonLogFormat
	ColouredTextLogFormat
)

var logFormat = map[string]Format{
	"text":         TextLogFormat,
	"json":         JsonLogFormat,
	"colouredText": ColouredTextLogFormat,
}

type Format int

func (f Format) String() string {
	return [...]string{"text", "json", "colouredText"}[f]
}

func ParseFormat(f string) (Format, error) {
	var ok bool
	var format Format
	if format, ok = logFormat[strings.ToLower(f)]; !ok {
		return JsonLogFormat, fmt.Errorf("unrecognized log format %q", f)
	}
	return format, nil
}
