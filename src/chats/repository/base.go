package repository

import "time"

type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type SortRequest struct {
	Field     string
	Direction string
}

type PagingRequest struct {
	Size   int
	Index  int
	SortBy []SortRequest
}

type PagingResponse struct {
	Total int
	Index int
}
