package transform

// Operation represents a single DOM transformation
type Operation struct {
	Type     string `json:"type"`
	Selector string `json:"selector"`
	Value    string `json:"value"`
	Property string `json:"property,omitempty"`
	Priority int    `json:"priority"`
}

// TransformSpec represents the full transformation specification
type TransformSpec struct {
	Version          string      `json:"version"`
	ExperimentID     string      `json:"experiment_id"`
	VariantID        string      `json:"variant_id"`
	VariantKey       string      `json:"variant_key"`
	Operations       []Operation `json:"operations"`
	TTL              int         `json:"ttl"`
	CacheKey         string      `json:"cache_key"`
	ExperimentVersion string     `json:"experiment_version,omitempty"`
}

// Variant represents an experiment variant
type Variant struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	IsControl         bool    `json:"is_control"`
	TrafficAllocation float64 `json:"traffic_allocation"`
}

// AssignmentRequest for variant assignment
type AssignmentRequest struct {
	ExperimentID string `json:"experiment_id"`
	UserID       string `json:"user_id,omitempty"`
}

// AssignmentResponse from variant assignment
type AssignmentResponse struct {
	VariantID  string `json:"variant_id"`
	VariantKey string `json:"variant_key"`
}

const (
	// Operation types
	OpSetText  = "setText"
	OpSetStyle = "setStyle"
	OpSetAttr  = "setAttr"
	OpSetHTML  = "setHTML"
	OpRemove   = "remove"
	OpHide     = "hide"
	OpShow     = "show"
)
