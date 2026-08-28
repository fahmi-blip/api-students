package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, u model.Student) (model.Student, error)
	Update(ctx context.Context, u model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

var kolomUrut = map[string]string{
	"id":    "id",
	"nim":   "nim",
	"name":  "name",
	"grade": "grade",
	"created_at": "created_at",
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func buildFilter(q model.ListQuery) (string, []any) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (nim ILIKE $%d OR name ILIKE $%d)",
			len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	if q.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", len(args)+1)
		args = append(args, *q.IsActive)
	}
	if q.MinGrade != nil {
		where += fmt.Sprintf(" AND grade >= $%d", len(args)+1)
		args = append(args, *q.MinGrade)
	}
	if q.MaxGrade != nil {
		where += fmt.Sprintf(" AND grade <= $%d", len(args)+1)
		args = append(args, *q.MaxGrade)
	}
	return where, args
}

func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.Student, int, error) {
	where, args := buildFilter(q)
	// 1) Hitung total sebelum dipenggal, untuk keperluan meta.
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung student: %w", err)
	}

	// 2) Ambil satu halaman saja. Penyaringan, pengurutan, dan pemenggalan // dikerjakan basis data, bukan oleh Go.
	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	sqlText := fmt.Sprintf(
		`SELECT id, nim, name, grade, is_active, created_at 
		 FROM students%s 
		 ORDER BY %s %s 
		 LIMIT $%d OFFSET $%d`,
		where, kolomUrut[q.Sort], arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var u model.Student
		if err := rows.Scan(&u.ID, &u.Nim, &u.Name, &u.Grade,
			&u.IsActive, &u.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		hasil = append(hasil, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}
	return hasil, total, nil
}

func (r *studentPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Student, error) {
	var u model.Student

	err := r.pool.QueryRow(ctx,
		`SELECT id, nim, name, grade, is_active
		FROM students WHERE id = $1`, id,
	).Scan(&u.ID, &u.Nim, &u.Name, &u.Grade, &u.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}
	return u, nil
}

func (r *studentPostgresRepository) Create(
	ctx context.Context, u model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active)
	VALUES ($1, $2, $3, $4)
	RETURNING id`,
		u.Nim, u.Name, u.Grade, u.IsActive,
	).Scan(&u.ID)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}
	return u, nil
}

func (r *studentPostgresRepository) Update(
	ctx context.Context, u model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students
		 SET name = $1, grade = $2, is_active = $3
		 WHERE id = $4
		 RETURNING id, nim, name, grade, is_active`,
		u.Name, u.Grade, u.IsActive, u.ID,
	).Scan(&u.ID, &u.Nim, &u.Name, &u.Grade, &u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}

	return u, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
