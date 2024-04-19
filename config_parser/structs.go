package config_parser

type Project struct {
	Path  string `json:"path"`
	Start string `json:"start"`
	Name  string `json:"name"`
	Ide   string `json:"ide"`
}

type Context struct {
	Name     string    `json:"name"`
	Projects []Project `json:"projects"`
}

type Config struct {
	Contexts []Context `json:"contexts"`
}
