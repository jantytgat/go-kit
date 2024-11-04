package semver

import "testing"

var validMetadataTests = []struct {
	name     string
	metadata Metadata
	commit   string
	date     string
	err      bool
}{
	{
		name:     "simple",
		metadata: "metadata.20101112",
		commit:   "metadata",
		date:     "20101112",
		err:      false,
	},
}

func TestSplitMetadata(t *testing.T) {
	for _, tt := range validMetadataTests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := SplitMetadata(tt.metadata)
			if (err != nil) != tt.err {
				t.Errorf("SplitMetadata() error = %v, wantErr %v", err, tt.err)
				return
			}
			if got != tt.commit {
				t.Errorf("SplitMetadata() got = %v, want %v", got, tt.commit)
			}
			if got1 != tt.date {
				t.Errorf("SplitMetadata() got1 = %v, want %v", got1, tt.date)
			}
		})
	}
}
