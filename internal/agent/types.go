package agent

type MultiModalResourceType string

const (
	MultiModalResourceTypeText  MultiModalResourceType = "text"
	MultiModalResourceTypeImage MultiModalResourceType = "image"
	MultiModalResourceTypeAudio MultiModalResourceType = "audio"
	MultiModalResourceTypeVideo MultiModalResourceType = "video"
	MultiModalResourceTypeFile  MultiModalResourceType = "file"
)

type MultiModalImageDetail string

const (
	MultiModalImageDetailHigh MultiModalImageDetail = "high"
	MultiModalImageDetailLow  MultiModalImageDetail = "low"
	MultiModalImageDetailAuto MultiModalImageDetail = "auto"
)

type MultiModalResource struct {
	Type       MultiModalResourceType `json:"type"`
	Text       string                 `json:"text,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Base64Data string                 `json:"base64_data,omitempty"`
	MIMEType   string                 `json:"mime_type,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Detail     MultiModalImageDetail  `json:"detail,omitempty"`
	Extra      map[string]any         `json:"extra,omitempty"`
}

type ChatRequest struct {
	Content             string
	MultiModalResources []MultiModalResource
	IsStream            bool
}
