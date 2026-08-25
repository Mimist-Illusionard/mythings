package models

type Tag struct {
	id   int64
	Name string
}

func NewTag(name string) *Tag {
	return &Tag{Name: name}
}

func (t *Tag) ID() int64 {
	return t.id
}
