package model

// Requests

type CreateProjectRequest struct {
	Name      string            `json:"name"`
	Logo      []byte            `json:"logo"`
	Team      map[string]string `json:"team"`
	ProjectID string            `json:"project_id"`
}

type AddProjectMemberRequest struct {
	Team      map[string]string `json:"team"`
	ProjectId string            `json:"project_id"`
}

type UpdateProjectMemberRequest struct {
	Team      map[string]string `json:"team"`
	ProjectId string            `json:"project_id"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ProjectResponse struct {
	Name      string            `json:"name"`
	Logo      []byte            `json:"logo"`
	Team      map[string]string `json:"team"`
	ProjectID string            `json:"project_id"`
}
