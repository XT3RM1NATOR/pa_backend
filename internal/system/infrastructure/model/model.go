package model

type Project struct {
	Name      string `bson:"name"`
	Logo      []byte
	Team      map[string]string
	ProjectID string `bson:"project_id"`
}
