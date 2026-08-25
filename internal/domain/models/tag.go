package models

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewTag(name string) *Tag {
	return &Tag{Name: name}
}
