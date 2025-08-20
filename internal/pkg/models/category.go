package models

type CategoryRequest struct {
	Query      string   `json:"query"`
	Categories []string `json:"categories"`
}

type FunctionRequest struct {
	Query     string   `json:"query"`
	Functions []string `json:"functions"`
}
