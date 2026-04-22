package handler

type GitHubRequest struct {
	Url    string `json:"url" example:"https://github.com/G97-TECH-MKT/my-repo/pull/42"`
	Action string `json:"action" example:"both" enums:"approve,merge,both"`
} // @name GitHubRequest

type GitHubResponse struct {
	Message string `json:"message"`
	Status  string `json:"status" enums:"success,error"`
}

type successResponse struct {
	Message string `json:"message" example:"Pull request approved and merged"`
	Status  string `json:"status" example:"success"`
} // @name SuccessResponse

type errorResponse struct {
	Message string `json:"message" example:"Invalid action"`
	Status  string `json:"status" example:"error"`
} // @name ErrorResponse
