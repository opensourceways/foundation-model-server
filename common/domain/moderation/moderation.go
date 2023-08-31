package moderation

type Moderation interface {
	CheckText(string) error
}
