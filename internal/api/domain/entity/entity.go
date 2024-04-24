package entity

type HelpDeskArticle struct {
	ArticleId int `bson:"_id"`
	ViewCount int `bson:"view_count"`
}
