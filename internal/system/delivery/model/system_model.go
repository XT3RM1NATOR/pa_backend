package model

// Requests

type CreateWorkspaceRequest struct {
	Name        string            `json:"name"`
	Logo        []byte            `json:"logo"`
	Team        map[string]string `json:"team"`
	Teams       []string          `json:"teams"`
	WorkspaceId string            `json:"workspace_id"`
}

type AddTeamMembersRequest struct {
	TeamName    string `json:"team_name"`
	WorkspaceId string `json:"workspace_id"`
	Member      string `json:"member"`
}

type AddWorkspaceMemberRequest struct {
	Team        map[string]string `json:"team"`
	WorkspaceId string            `json:"Workspace_id"`
}

type UpdateWorkspaceMemberRequest struct {
	Team        map[string]string `json:"team"`
	WorkspaceId string            `json:"workspace_id"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Logo        []byte `json:"logo"`
	WorkspaceID string `json:"workspace_id"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type WorkspaceResponse struct {
	Name        string `json:"name"`
	Logo        []byte `json:"logo"`
	WorkspaceID string `json:"workspace_id"`
}

type UserResponse struct {
	Email    string `json:"email"`
	FullName string `json:"name"`
	Role     string `json:"role"`
	Logo     []byte `json:"logo"`
}
