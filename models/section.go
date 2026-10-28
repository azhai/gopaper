package models

type SectionInput struct {
	Name      string `json:"name"`
	DirPath   string `json:"dirPath"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	NavOrder  int    `json:"navOrder"`
	NavHide   bool   `json:"navHide"`
	Layout    string `json:"layout"`
	SortOrder string `json:"sortOrder"`
}
