package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

func ReqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

var allowedSort = map[string]bool{
	"id": true, "name": true, "grade": true, "nim": true,
}

// parseListQuery membaca query string dan memberi nilai bawaan yang aman.
// Aturan pentingnya: masukan dari klien tidak pernah dipercaya begitu saja.
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("min_grade"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MinGrade = &v
		}
	}
	if raw := c.Query("max_grade"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MaxGrade = &v
		}
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	return q
}

func ParseStudentCursorQuery(c *fiber.Ctx) (model.StudentCursorQuery, error) {
	query := model.StudentCursorQuery{
		Search: strings.TrimSpace(c.Query("search")),
		Limit:  c.QueryInt("limit", 10),
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if raw := c.Query("is_active"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return model.StudentCursorQuery{}, BadRequest("nilai is_active harus true atau false")
		}
		query.IsActive = &value
	}
	if raw := c.Query("cursor"); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return model.StudentCursorQuery{}, BadRequest("cursor tidak valid")
		}
		query.After = &cursor
	}
	return query, nil
}

func ParseUserCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	query := model.CursorQuery{
		Search: strings.TrimSpace(c.Query("search")),
		Limit:  c.QueryInt("limit", 10),
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if raw := c.Query("is_active"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return model.CursorQuery{}, BadRequest("nilai is_active harus true atau false")
		}
		query.IsActive = &value
	}
	if raw := c.Query("cursor"); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return model.CursorQuery{}, BadRequest("cursor tidak valid")
		}
		query.After = &cursor
	}
	return query, nil
}
