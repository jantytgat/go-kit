package flagzog

import (
	"fmt"

	"github.com/Oudwins/zog"
	"github.com/spf13/pflag"
)

func NewBoolFlag(name string, schema *zog.BoolSchema[bool], usage string) BoolFlag {
	return BoolFlag{
		name:   name,
		schema: schema,
		usage:  usage,
	}
}

type BoolFlag struct {
	name   string
	schema *zog.BoolSchema[bool]
	usage  string
	Value  bool
}

func (f BoolFlag) Name() string {
	return f.name
}

func (f BoolFlag) Usage() string {
	return f.usage
}

func (f BoolFlag) Validate() ([]string, error) {
	var messages []string
	if issues := f.schema.Validate(&f.Value); issues != nil {
		for _, issue := range issues {
			messages = append(messages, issue.Message)
		}
		return messages, fmt.Errorf("validation failed for flag '%s' with value '%t'", f.Name(), f.Value)
	}
	return messages, nil
}

func (f BoolFlag) AddToCommandFlags(flagset *pflag.FlagSet, shorthand string, value interface{}) {
	flagset.BoolVarP(&f.Value, f.Name(), shorthand, value.(bool), f.usage)
}

func NewInt64Flag(name string, schema *zog.NumberSchema[int64], usage string) Int64Flag {
	return Int64Flag{
		name:   name,
		schema: schema,
		usage:  usage,
	}
}

type Int64Flag struct {
	name   string
	schema *zog.NumberSchema[int64]
	usage  string
	Value  int64
}

func (f Int64Flag) Name() string {
	return f.name
}

func (f Int64Flag) Usage() string {
	return f.usage
}

func (f Int64Flag) Validate() ([]string, error) {
	var messages []string
	if issues := f.schema.Validate(&f.Value); issues != nil {
		for _, issue := range issues {
			messages = append(messages, issue.Message)
		}
		return messages, fmt.Errorf("validation failed for flag '%s' with value '%d'", f.Name(), f.Value)
	}
	return messages, nil
}

func (f Int64Flag) AddToCommandFlags(flagset *pflag.FlagSet, shorthand string, value interface{}) {
	flagset.Int64VarP(&f.Value, f.Name(), shorthand, value.(int64), f.usage)
}

func NewStringFlag(name string, schema *zog.StringSchema[string], usage string) StringFlag {
	return StringFlag{
		name:   name,
		schema: schema,
		usage:  usage,
	}
}

type StringFlag struct {
	name   string
	schema *zog.StringSchema[string]
	usage  string
	Value  string
}

func (f StringFlag) Name() string {
	return f.name
}

func (f StringFlag) Usage() string {
	return f.usage
}

func (f StringFlag) Validate() ([]string, error) {
	var messages []string
	if issues := f.schema.Validate(&f.Value); issues != nil {
		for _, issue := range issues {
			messages = append(messages, issue.Message)
		}
		return messages, fmt.Errorf("validation failed for flag '%s' with value '%s'", f.Name(), f.Value)
	}
	return messages, nil
}

func (f StringFlag) AddToCommandFlags(flagset *pflag.FlagSet, shorthand string, value interface{}) {
	flagset.StringVarP(&f.Value, f.Name(), shorthand, value.(string), f.usage)
}
