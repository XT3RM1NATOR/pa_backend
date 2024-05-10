package infrastructureModel

type Workspace struct {
	Name        string `bson:"name"`
	Logo        []byte
	Team        map[string]string
	WorkspaceId string `bson:"workspace_id"`
}

type User struct {
	Email    string `bson:"email"`
	FullName string `bson:"name"`
	Role     string
	Logo     []byte
}
