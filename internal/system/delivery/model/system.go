package model

// Requests

type CreateProjectRequest struct {
	Name      string   `json:"name"`
	Logo      []byte   `json:"logo"`
	Team      []string `json:"team"`
	ProjectID string   `json:"project_id"`
}

type LeaveProjectRequest struct {
	ProjectID string `json:"project_id"`
}

type GetProjectByIdRequest struct {
	ProjectID string `json:"project_id"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
