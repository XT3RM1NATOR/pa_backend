package model

type Project struct {
	Name      string `bson:"name"`
	Logo      []byte
	Team      map[string]string
	ProjectID string `bson:"project_id"`
}

type User struct {
	Email    string `bson:"email"`
	FullName string `bson:"name"`
	Role     string
	Logo     []byte
}
