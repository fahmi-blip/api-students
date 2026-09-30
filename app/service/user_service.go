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
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func (h *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseStudentCursorQuery(c)
	if err != nil {
		return err
	}

	students, err := h.repo.FindAfterCursor(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}
	hasMore := len(students) > q.Limit
	if hasMore {
		students = students[:q.Limit]
	}
	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, students)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(students) > 0 {
		last := students[len(students)-1]
		meta.NextCursor = helper.EncodeCursor(last.CreatedAt, last.ID)
	}
	return helper.SuccessCursor(c, "daftar student berhasil diambil", students, meta)
}

func (h *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(err, "user")
	}
	if !CanAccessStudent(current, student.OwnerID, h.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data user lain")
	}

	return helper.Ok(c, fiber.StatusOK, "user ditemukan", student)
}

func (h *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreatedStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Nim = strings.TrimSpace(req.Nim)
	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	newStudent, err := h.repo.Create(ctx, model.Student{
		Nim:      req.Nim,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  current.UserID,
	})

	if err != nil {
		return terjemahkanError(err, "gagal menyimpan student")
	}

	return helper.Created(c, "student berhasil dibuat", newStudent,
		"/api/v1/students/"+strconv.Itoa(newStudent.ID),
	)
}
func (h *StudentService) GetWithPrestation(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	student, err := h.repo.FindByIDWithPrestasi(ctx, id)
	if err != nil {
		return terjemahkanError(err, "gagal mengambil data student dengan prestasi")
	}
	if !CanAccessStudent(current, student.OwnerID, h.perms, "student:read:any") {
		return helper.Forbidden("tidak berhak mengakses data student ini")
	}
	return helper.Ok(c, fiber.StatusOK, "student dengan prestasi ditemukan", student)
}

func (h *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	existing, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(err, "gagal mengambil data student")
	}

	if !CanAccessStudent(current, existing.OwnerID, h.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student ini")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	result, err := h.repo.Update(ctx, model.Student{
		ID:       id,
		Name:     strings.TrimSpace(req.Name),
		Grade:    req.Grade,
		IsActive: req.IsActive,
	})

	if err != nil {
		return terjemahkanError(err,
			"gagal memperbarui student")
	}
	return helper.Ok(c, fiber.StatusOK, "student berhasil diganti seluruhnya", result)
}

func (h *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}
	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		return helper.Validation(errs)
	}

	saatIni, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return terjemahkanError(err, "gagal mengambil data student")
	}

	if !CanAccessStudent(current, saatIni.OwnerID, h.perms, "student:update:any") {
		return helper.Forbidden("tidak berhak mengubah data student ini")
	}

	updated := ApplyPatch(saatIni, req)

	result, err := h.repo.Update(ctx, updated)
	if err != nil {
		return terjemahkanError(err, "gagal memperbarui student")
	}
	return helper.Ok(c, fiber.StatusOK, "student berhasil diperbarui sebagian", result)
}

func (h *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.ReqCtx(c)
	defer cancel()

	id, valid := helper.ParamID(c)

	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		return terjemahkanError(err, "gagal menghapus user")
	}

	return helper.NoContent(c)
}

func terjemahkanError(err error, pesanUmum string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(pesanUmum + " student tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("nim sudah terdaftar")
	default:
		return helper.Internal(err)
	}
}