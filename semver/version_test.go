package semver

import (
	"reflect"
	"testing"
)

var validVersionTests = []struct {
	name     string
	input    string
	version  Version
	inputErr bool
	commit   string
	date     string
	release  string
	full     string
	number   string
}{
	{
		name:  "stable",
		input: "0.1.0",
		version: Version{
			Major: 0,
			Minor: 1,
		},
		inputErr: false,
		commit:   "",
		date:     "",
		release:  "stable",
		full:     "0.1.0",
		number:   "0.1.0",
	},
	{
		name:  "stable.1",
		input: "0.1.1",
		version: Version{
			Major: 0,
			Minor: 1,
			Patch: 1,
		},
		inputErr: false,
		commit:   "",
		date:     "",
		release:  "stable",
		full:     "0.1.1",
		number:   "0.1.1",
	},
	{
		name:  "stable.1+metadata",
		input: "0.1.1+metadata",
		version: Version{
			Major:    0,
			Minor:    1,
			Patch:    1,
			Metadata: "metadata",
		},
		inputErr: false,
		commit:   "metadata",
		date:     "metadata",
		release:  "stable",
		full:     "0.1.1+metadata",
		number:   "0.1.1",
	},
	{
		name:  "stable.1+metadata.date",
		input: "0.1.1+metadata.20101112",
		version: Version{
			Major:    0,
			Minor:    1,
			Patch:    1,
			Metadata: "metadata.20101112",
		},
		inputErr: false,
		commit:   "metadata",
		date:     "20101112",
		release:  "stable",
		full:     "0.1.1+metadata.20101112",
		number:   "0.1.1",
	},
	{
		name:  "alpha",
		input: "0.1.1-alpha",
		version: Version{
			Major:      0,
			Minor:      1,
			Patch:      1,
			PreRelease: "alpha",
		},
		inputErr: false,
		commit:   "",
		release:  "alpha",
		full:     "0.1.1-alpha",
		number:   "0.1.1",
	},
	{
		name:  "alpha.1",
		input: "0.1.1-alpha.1",
		version: Version{
			Major:      0,
			Minor:      1,
			Patch:      1,
			PreRelease: "alpha.1",
		},
		inputErr: false,
		commit:   "",
		release:  "alpha.1",
		full:     "0.1.1-alpha.1",
		number:   "0.1.1",
	},
}

func TestParse(t *testing.T) {
	for _, tt := range validVersionTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.inputErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.inputErr)
				return
			}
			if !reflect.DeepEqual(got, tt.version) {
				t.Errorf("Parse() got = %v, want %v", got, tt.version)
			}
		})
	}
}

func TestVersion_Commit(t *testing.T) {
	for _, tt := range validVersionTests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:      tt.version.Major,
				Minor:      tt.version.Minor,
				Patch:      tt.version.Patch,
				PreRelease: tt.version.PreRelease,
				Metadata:   tt.version.Metadata,
			}
			if got := v.Commit(); got != tt.commit {
				t.Errorf("Commit() = %v, want %v", got, tt.commit)
			}
		})
	}
}

func TestVersion_Date(t *testing.T) {
	for _, tt := range validVersionTests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:      tt.version.Major,
				Minor:      tt.version.Minor,
				Patch:      tt.version.Patch,
				PreRelease: tt.version.PreRelease,
				Metadata:   tt.version.Metadata,
			}
			if got := v.Date(); got != tt.date {
				t.Errorf("Date() = %v, want %v", got, tt.date)
			}
		})
	}
}

func TestVersion_Release(t *testing.T) {
	for _, tt := range validVersionTests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:      tt.version.Major,
				Minor:      tt.version.Minor,
				Patch:      tt.version.Patch,
				PreRelease: tt.version.PreRelease,
				Metadata:   tt.version.Metadata,
			}
			if got := v.Release(); got != tt.release {
				t.Errorf("Release() = %v, want %v", got, tt.release)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	for _, tt := range validVersionTests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:      tt.version.Major,
				Minor:      tt.version.Minor,
				Patch:      tt.version.Patch,
				PreRelease: tt.version.PreRelease,
				Metadata:   tt.version.Metadata,
			}
			if got := v.String(); got != tt.full {
				t.Errorf("String() = %v, want %v", got, tt.full)
			}
		})
	}
}

func TestVersion_VersionNumber(t *testing.T) {
	for _, tt := range validVersionTests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:      tt.version.Major,
				Minor:      tt.version.Minor,
				Patch:      tt.version.Patch,
				PreRelease: tt.version.PreRelease,
				Metadata:   tt.version.Metadata,
			}
			if got := v.Number(); got != tt.number {
				t.Errorf("Number() = %v, want %v", got, tt.number)
			}
		})
	}
}
