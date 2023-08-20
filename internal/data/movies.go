package data

import (
	"database/sql"
	"time"

	"github.com/Alphasxd/greenlight/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, movie *Movie) {

	// 检查 Title 字段
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	// 检查 Year 字段
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// 检查 Runtime 字段
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	// 检查 Genres 字段
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// Insert 方法将一个新的电影添加到 movies 数据表中。
func (m MovieModel) Insert(movie *Movie) error {
	return nil
}

// Get 方法返回指定 ID 的电影。
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update 方法用来更新指定 ID 的电影信息。
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// Delete 方法用来删除指定 ID 的电影。
func (m MovieModel) Delete(id int64) error {
	return nil
}
