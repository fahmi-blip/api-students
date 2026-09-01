package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (h *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	students, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c,fiber.StatusInternalServerError,"gagal mengambil data student",
		)
	}

	return helper.OkList(c,"daftar student berhasil diambil",students,&model.Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPage:  CountTotalPages(total, q.Limit),
		},
	)
}

func (h *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c,err,"gagal mengambil data student",)
	}

	return helper.Ok(c, fiber.StatusOK,"user ditemukan", student)
}

func (h *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	var req model.CreatedStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c,fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",)
	}

	req.Nim = strings.TrimSpace(req.Nim)
	req.Name = strings.TrimSpace(req.Name)

	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	newStudent, err := h.repo.Create(ctx, model.Student{
		Nim: req.Nim,
		Name: req.Name,
		Grade: req.Grade,
		IsActive: true,
	})

	if err != nil {
		return terjemahkanError(c,err,"gagal menyimpan student")
	}

	return helper.Created(c,"student berhasil dibuat",newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID),
	)
}

func (h *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c,fiber.StatusBadRequest,
			"id harus berupa angka positif")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c,fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",)
	}

	if errs := ValidateReplace(req); len(errs) > 0{
		return helper.FailValidation(c, errs)
	}

	result, err := h.repo.Update(ctx, model.Student{
		ID:       id,
		Name: strings.TrimSpace(req.Name),
		Grade: req.Grade,
		IsActive: req.IsActive,
	})
	
	if err != nil {
		return terjemahkanError(c,err,
			"gagal memperbarui student",)
	}
	return helper.Ok(c, fiber.StatusOK, "student berhasil diganti seluruhnya", result)
}

func (h *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.Fail(c,fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c,fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	if IsEmptyPatch(req){
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	saatIni, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(c,err,"gagal mengambil data student")
	}

	updated, errs := ApplyPatch(saatIni, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := h.repo.Update(ctx, updated)
	if err != nil {
		return terjemahkanError(c,err,"gagal memperbarui student",)
	}
	return helper.Ok(c, fiber.StatusOK,"student berhasil diperbarui sebagian",result)
}

func (h *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		return terjemahkanError(c,err,"gagal menghapus user",)
	}

	return helper.NoContent(c)
}

func terjemahkanError(c *fiber.Ctx,err error,pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")

	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "nim sudah terdaftar")

	default:
		return helper.Fail(c, fiber.StatusInternalServerError, pesanUmum)
	}
}