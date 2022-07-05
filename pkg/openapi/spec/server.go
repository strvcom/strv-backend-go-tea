package spec

type Server struct {
	URL         string           `json:"url,omitempty"`
	Description string           `json:"description,omitempty"`
	Variables   []ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}
