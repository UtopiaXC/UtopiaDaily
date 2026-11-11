package classification

import "context"

type InputItem struct {
	Fingerprint string
	Title       string
	Summary     string
	Content     string
	SourceName  string
}

type TagAssignment struct {
	GroupName  string
	TagName    string
	Confidence float64
}

type Result struct {
	Fingerprint string
	Tags        []TagAssignment
}

type SlotStatus struct {
	Provider      string `json:"provider"`
	SlotReady     bool   `json:"slot_ready"`
	TagGroupCount int64  `json:"tag_group_count"`
	TagCount      int64  `json:"tag_count"`
}

type Classifier interface {
	Provider() string
	ClassifyBatch(ctx context.Context, items []InputItem) ([]Result, error)
}
