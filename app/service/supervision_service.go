package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

var errSameSupervisor = errors.New("pembimbing tujuan sama dengan pembimbing saat ini")

type SupervisionService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

func NewSupervisionService(s repository.SupervisorRepository, tx *repository.TxManager) *SupervisionService {
	return &SupervisionService{supervisors: s, tx: tx}
}

// POST /supervisors/:id/assign
func (s *SupervisionService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	supID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.AssignRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var result model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx) // seluruh query lewat repo, bukan s.supervisors

		// 1. Kunci baris mahasiswa: penetapan serentak untuk mahasiswa
		//    yang sama harus antre.
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}
		// 2. Aman diperiksa karena langkah 1.
		if _, err := repo.CurrentSupervisor(ctx, req.StudentID); err == nil {
			return repository.ErrAlreadySupervised
		} else if !errors.Is(err, repository.ErrNotSupervised) {
			return err
		}
		// 3. Ambil satu slot secara atomik (kuota).
		sup, err := repo.IncrementSupervised(ctx, supID)
		if err != nil {
			return err
		}
		result = sup
		// 4. Catat. PRIMARY KEY menjadi jaring terakhir; error di sini
		//    membatalkan juga langkah 3.
		return repo.InsertSupervision(ctx, req.StudentID, supID)
	})
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing berhasil ditetapkan", result)
}

// POST /supervisions/switch
func (s *SupervisionService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var req model.SwitchSupervisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	var result model.Supervisor
	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}
		from, err := repo.CurrentSupervisor(ctx, req.StudentID)
		if err != nil {
			return err
		}
		if from == req.ToSupervisorID {
			return errSameSupervisor
		}
		// Kunci kedua pembimbing dalam urutan id menaik: bebas deadlock.
		if err := repo.LockSupervisors(ctx, from, req.ToSupervisorID); err != nil {
			return err
		}
		if err := repo.DecrementSupervised(ctx, from); err != nil {
			return err
		}
		// Bila pembimbing baru penuh, error ini me-rollback pengurangan di atas:
		// mahasiswa tetap dengan pembimbing lama.
		sup, err := repo.IncrementSupervised(ctx, req.ToSupervisorID)
		if err != nil {
			return err
		}
		result = sup
		return repo.MoveSupervision(ctx, req.StudentID, req.ToSupervisorID)
	})
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Success(c, fiber.StatusOK, "pembimbing berhasil diganti", result)
}

// PATCH /supervisors/:id (satu UPDATE, tanpa WithTx)
func (s *SupervisionService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}
	var req model.UpdateSupervisorQuotaRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	updated, err := s.supervisors.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		return translateSupervisionError(err)
	}
	return helper.Success(c, fiber.StatusOK, "kuota berhasil diubah", updated)
}

func translateSupervisionError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(fiber.StatusConflict, "SUPERVISOR_FULL", "kuota pembimbing sudah penuh")
	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(fiber.StatusConflict, "ALREADY_SUPERVISED", "mahasiswa sudah memiliki pembimbing")
	case errors.Is(err, repository.ErrNotSupervised):
		return helper.NewError(fiber.StatusConflict, "NOT_SUPERVISED", "mahasiswa belum memiliki pembimbing")
	case errors.Is(err, errSameSupervisor):
		return helper.NewError(fiber.StatusConflict, "SAME_SUPERVISOR", "pembimbing tujuan sama dengan saat ini")
	case errors.Is(err, repository.ErrVersionConflict):
		return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT",
			"data sudah diubah pihak lain, muat ulang lalu coba lagi")
	case errors.Is(err, repository.ErrQuotaBelowFilled):
		return helper.NewError(fiber.StatusConflict, "QUOTA_BELOW_SUPERVISED",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa bimbingan")
	default:
		return helper.Internal(err)
	}
}