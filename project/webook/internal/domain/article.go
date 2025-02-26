package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
}

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

type ArticleStatus uint8

func (a ArticleStatus) ToUint8() uint8 {
	return uint8(a)
}

func (a ArticleStatus) NonPublished() bool {
	return a != ArticleStatusPublished
}

func (a ArticleStatus) String() string {
	switch a {
	case ArticleStatusUnpublished:
		return "Unpublished"
	case ArticleStatusPublished:
		return "Published"
	case ArticleStatusPrivate:
		return "Private"
	default:
		return "Unknown"

	}
}

type Author struct {
	Id   int64
	Name string
}
