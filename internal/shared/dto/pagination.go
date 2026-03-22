package dto

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type PaginationRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Items    interface{}    `json:"items"`
	Metadata PaginationMeta `json:"metadata"`
}
