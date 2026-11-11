package classification

import "context"

type NoopClassifier struct{}

func NewNoopClassifier() *NoopClassifier {
	return &NoopClassifier{}
}

func (n *NoopClassifier) Provider() string {
	return "slot"
}

func (n *NoopClassifier) ClassifyBatch(ctx context.Context, items []InputItem) ([]Result, error) {
	results := make([]Result, 0, len(items))
	for _, item := range items {
		results = append(results, Result{
			Fingerprint: item.Fingerprint,
			Tags:        nil,
		})
	}
	return results, nil
}
