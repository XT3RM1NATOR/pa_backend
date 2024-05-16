package entity

type HelpDeskArticle struct {
	ArticleId int      `bson:"_id"`
	ViewCount int      `bson:"view_count"`
	Language  Language `bson:"language"`
}

type Language string

const (
	English  Language = "en"
	Russian  Language = "ru"
	Uzbek    Language = "uz"
	UzbekCyr Language = "uz-uz"
)
