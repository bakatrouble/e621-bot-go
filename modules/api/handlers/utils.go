package handlers

type subsRequestBody struct {
	Subs []string `json:"subs" binding:"required"`
}
