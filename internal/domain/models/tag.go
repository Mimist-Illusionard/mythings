package models

type Tag struct {
	ID   int64
	Name string
}

func NewTag(name string) *Tag {
	return &Tag{Name: name}
}
