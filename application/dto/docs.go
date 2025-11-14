package dto

// DocItem represents a documentation item (file or folder)
type DocItem struct {
	Name     string
	Path     string
	IsFolder bool
	Children []DocItem
}
