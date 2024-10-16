package data 

import (
	"time"
)

type Movie struct {
	ID int64 `json:"id"`
	Title string `json:"title"`
	Year int32 `json:"year,omitempty"`
	Runtime Runtime `json:"runtime,omitempty"`
	Genre []string `json:"genre,omitempty"`
	Version int32 `json:"version"`
	CreatedAt time.Time `json:"-"`
}