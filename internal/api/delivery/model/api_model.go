package model

// Requests

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type HelpDeskArticleResponse struct {
	ArticleId int `json:"article_id"`
	ViewCount int `json:"view_count"`
}
