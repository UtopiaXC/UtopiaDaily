package classification

import "context"

const (
	TruncateStrategyTitleHeadTail = "title_head_tail"
	TruncateStrategyHead          = "head"
	TruncateStrategyTail          = "tail"

	DefaultMaxInputChars = 2000
)

type InputItem struct {
	ResultID    uint   `json:"result_id"`
	Fingerprint string `json:"fingerprint"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Content     string `json:"content"`
	SourceName  string `json:"source_name"`
}

type TagAssignment struct {
	GroupName  string  `json:"group_name"`
	TagName    string  `json:"tag_name"`
	TagID      uint    `json:"tag_id"`
	Confidence float64 `json:"confidence"`
	Provider   string  `json:"provider"`
	Reason     string  `json:"reason,omitempty"`
}

type Result struct {
	Fingerprint string
	Tags        []TagAssignment
}

type TagSpec struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Priority    int      `json:"priority"`
	Keywords    []string `json:"keywords"`
}

type GroupSpec struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	SelectionMode string    `json:"selection_mode"`
	MaxResults    int       `json:"max_results"`
	Tags          []TagSpec `json:"tags"`
}

type ProfileSpec struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Groups      []GroupSpec `json:"groups"`
}

type EngineOptions struct {
	Provider         string      `json:"provider"`
	ModelName        string      `json:"model_name"`
	Confidence       int         `json:"confidence"`
	MaxInputChars    int         `json:"max_input_chars"`
	TruncateStrategy string      `json:"truncate_strategy"`
	LLMProtocol      string      `json:"llm_protocol"`
	LLMURL           string      `json:"llm_url"`
	LLMAPIKey        string      `json:"llm_api_key"`
	LLMModel         string      `json:"llm_model"`
	LLMPrompt        string      `json:"llm_prompt"`
	Profile          ProfileSpec `json:"profile"`
}

type SlotStatus struct {
	Provider      string `json:"provider"`
	SlotReady     bool   `json:"slot_ready"`
	TagGroupCount int64  `json:"tag_group_count"`
	TagCount      int64  `json:"tag_count"`
	WorkerState   string `json:"worker_state"`
	PendingTasks  int64  `json:"pending_tasks"`
}

type PresetModelItem struct {
	Name      string `json:"name"`
	HFRepo    string `json:"hf_repo"`
	SizeAbout string `json:"size_about"`
	Desc      string `json:"desc"`
}

type PresetModelGroup struct {
	Group  string            `json:"group"`
	Models []PresetModelItem `json:"models"`
}

type LocalONNXModel struct {
	Value     string `json:"value"`
	Label     string `json:"label"`
	HFRepo    string `json:"hf_repo,omitempty"`
	Desc      string `json:"desc,omitempty"`
	SizeAbout string `json:"size_about,omitempty"`
	SizeBytes int64  `json:"size_bytes"`
	Ready     bool   `json:"ready"`
}

type ONNXEnvStatus struct {
	Ready           bool   `json:"ready"`
	PythonPath      string `json:"python_path"`
	RuntimeDir      string `json:"runtime_dir"`
	ModelsDir       string `json:"models_dir"`
	LocalModelCount int    `json:"local_model_count"`
}

type OllamaModelInfo struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at,omitempty"`
	Family     string `json:"family,omitempty"`
	ParamSize  string `json:"parameter_size,omitempty"`
	QuantLevel string `json:"quantization_level,omitempty"`
}

type ProgressUpdate struct {
	Stage          string  `json:"stage"`
	Message        string  `json:"message,omitempty"`
	CurrentItem    string  `json:"current_item,omitempty"`
	CompletedBytes int64   `json:"completed_bytes,omitempty"`
	TotalBytes     int64   `json:"total_bytes,omitempty"`
	Percent        float64 `json:"percent"`
	Indeterminate  bool    `json:"indeterminate"`
}

type Classifier interface {
	Provider() string
	ClassifyBatch(ctx context.Context, items []InputItem) ([]Result, error)
}

type EngineClassifier interface {
	Provider() string
	ClassifyItem(ctx context.Context, item InputItem, opts EngineOptions) ([]TagAssignment, error)
}
