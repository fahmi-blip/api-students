package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"api-students/app/model"
)

var (
	ErrSupervisorFull    = errors.New("kuota pembimbing sudah penuh")
	ErrAlreadySupervised = errors.New("mahasiswa sudah memiliki pembimbing")
	ErrNotSupervised     = errors.New("mahasiswa belum memiliki pembimbing")
)

const supervisorColumns = `id, name, quota, supervised_count, version, created_at`

type SupervisorRepository interface {
	WithTx(tx DBTX) SupervisorRepository
	FindByID(ctx context.Context, id int) (model.Supervisor, error)
	LockStudent(ctx context.Context, studentID int) error
	CurrentSupervisor(ctx context.Context, studentID int) (int, error)
	LockSupervisors(ctx context.Context, ids ...int) error
	IncrementSupervised(ctx context.Context, id int) (model.Supervisor, error)
	DecrementSupervised(ctx context.Context, id int) error
	InsertSupervision(ctx context.Context, studentID, supervisorID int) error
	MoveSupervision(ctx context.Context, studentID, toSupervisorID int) error
	UpdateQuota(ctx context.Context, id, quota, version int) (model.Supervisor, error)
}

type supervisorPostgresRepository struct{ db DBTX }

func NewSupervisorRepository(db DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: db}
}

func (r *supervisorPostgresRepository) WithTx(tx DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: tx}
}

func scanSupervisor(row pgx.Row) (model.Supervisor, error) {
	var s model.Supervisor
	err := row.Scan(&s.ID, &s.Name, &s.Quota, &s.SupervisedCount, &s.Version, &s.CreatedAt)
	return s, err
}

func (r *supervisorPostgresRepository) FindByID(ctx context.Context, id int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		"SELECT "+supervisorColumns+" FROM supervisors WHERE id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, ErrNotFound
	}
	if err != nil {
		return model.Supervisor{}, fmt.Errorf("mengambil pembimbing: %w", err)
	}
	return s, nil
}

// LockStudent: pola LockStudent Modul 9. WAJIB memakai FOR UPDATE.
func (r *supervisorPostgresRepository) LockStudent(ctx context.Context, studentID int) error {
	var id int
	err := r.db.QueryRow(ctx,
		"SELECT id FROM students WHERE id = $1 FOR UPDATE", studentID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("mengunci mahasiswa: %w", err)
	}
	return nil
}

func (r *supervisorPostgresRepository) CurrentSupervisor(ctx context.Context, studentID int) (int, error) {
	var sid int
	err := r.db.QueryRow(ctx,
		"SELECT supervisor_id FROM supervisions WHERE student_id = $1", studentID).Scan(&sid)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotSupervised
	}
	if err != nil {
		return 0, fmt.Errorf("membaca pembimbing: %w", err)
	}
	return sid, nil
}

// LockSupervisors: pola LockCourses. Urutan id menaik mencegah deadlock.
func (r *supervisorPostgresRepository) LockSupervisors(ctx context.Context, ids ...int) error {
	rows, err := r.db.Query(ctx,
		"SELECT id FROM supervisors WHERE id = ANY($1) ORDER BY id FOR UPDATE", ids)
	if err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	defer rows.Close()
	found := 0
	for rows.Next() {
		found++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	if found != len(ids) {
		return ErrNotFound
	}
	return nil
}

// IncrementSupervised: UPDATE bersyarat atomik, pola IncrementEnrolled.
func (r *supervisorPostgresRepository) IncrementSupervised(ctx context.Context, id int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET supervised_count = supervised_count + 1
		 WHERE id = $1 AND supervised_count < quota
		 RETURNING `+supervisorColumns, id))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("menambah bimbingan: %w", err)
	}
	if _, findErr := r.FindByID(ctx, id); findErr != nil {
		return model.Supervisor{}, findErr
	}
	return model.Supervisor{}, ErrSupervisorFull
}

func (r *supervisorPostgresRepository) DecrementSupervised(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE supervisors SET supervised_count = supervised_count - 1
		 WHERE id = $1 AND supervised_count > 0`, id)
	if err != nil {
		return fmt.Errorf("mengurangi bimbingan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *supervisorPostgresRepository) InsertSupervision(ctx context.Context, studentID, supervisorID int) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO supervisions (student_id, supervisor_id) VALUES ($1, $2)", studentID, supervisorID)
	if err != nil {
		if pgErrorCode(err) == codeUniqueViolation {
			return ErrAlreadySupervised
		}
		return fmt.Errorf("mencatat pembimbing: %w", err)
	}
	return nil
}

func (r *supervisorPostgresRepository) MoveSupervision(ctx context.Context, studentID, toSupervisorID int) error {
	tag, err := r.db.Exec(ctx,
		"UPDATE supervisions SET supervisor_id = $2 WHERE student_id = $1", studentID, toSupervisorID)
	if err != nil {
		return fmt.Errorf("memindahkan pembimbing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotSupervised
	}
	return nil
}

// UpdateQuota: optimistic locking, pola Modul 9 (version naik 1 pada setiap perubahan).
func (r *supervisorPostgresRepository) UpdateQuota(ctx context.Context, id, quota, version int) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(ctx,
		`UPDATE supervisors SET quota = $1, version = version + 1
		 WHERE id = $2 AND version = $3 AND $1 >= supervised_count
		 RETURNING `+supervisorColumns, quota, id, version))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("mengubah kuota: %w", err)
	}
	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Supervisor{}, findErr
	}
	if current.Version != version {
		return model.Supervisor{}, ErrVersionConflict
	}
	return model.Supervisor{}, ErrQuotaBelowFilled
}