package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository // sesuaikan dengan nama di Modul 3
}

func NewPortfolioService(p repository.PortfolioRepository, s repository.StudentRepository) *PortfolioService {
	return &PortfolioService{portfolios: p, students: s}
}

// requireStudent: tidak ada FOREIGN KEY lintas database, jadi service
// yang harus bertanya ke PostgreSQL.
func (s *PortfolioService) studentID(c *fiber.Ctx) (int, error) {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, valid := helper.ParamID(c)
	if !valid {
		return 0, helper.BadRequest("id mahasiswa harus berupa angka positif")
	}
	if _, err := s.students.FindByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, helper.NotFound("mahasiswa tidak ditemukan")
		}
		return 0, translatePortfolioError(err)
	}
	return id, nil
}

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, err := s.studentID(c)
	if err != nil {
		return err
	}
	p, err := s.portfolios.FindByStudent(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("portofolio tidak ditemukan")
		}
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", p)
}

func (s *PortfolioService) Put(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, err := s.studentID(c)
	if err != nil {
		return err
	}
	var req model.PutPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	for i := range req.Projects {
		req.Projects[i].Technologies = NormalizeTags(req.Projects[i].Technologies)
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	p := model.Portfolio{StudentID: id, Headline: req.Headline, Bio: req.Bio}
	for _, pr := range req.Projects {
		p.Projects = append(p.Projects, model.Project{Title: pr.Title,
			Technologies: pr.Technologies, URL: pr.URL, Year: pr.Year})
	}
	for _, ce := range req.Certificates {
		t, _ := time.Parse("2006-01-02", ce.IssuedAt) // sudah lolos validasi
		p.Certificates = append(p.Certificates, model.Certificate{Name: ce.Name, Issuer: ce.Issuer, IssuedAt: t})
	}

	saved, err := s.portfolios.Upsert(ctx, p)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", saved)
}

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, err := s.studentID(c)
	if err != nil {
		return err
	}
	var req model.ProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Technologies = NormalizeTags(req.Technologies)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}
	updated, err := s.portfolios.AddProject(ctx, id, model.Project{
		Title: req.Title, Technologies: req.Technologies, URL: req.URL, Year: req.Year})
	if err != nil {
		return translatePortfolioError(err)
	}
	return helper.Success(c, fiber.StatusCreated, "proyek berhasil ditambahkan", updated)
}

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, err := s.studentID(c)
	if err != nil {
		return err
	}
	if err := s.portfolios.DeleteProject(ctx, id, c.Params("projectId")); err != nil {
		return translatePortfolioError(err)
	}
	return helper.NoContent(c)
}

func (s *PortfolioService) TopTechnologies(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	stats, err := s.portfolios.TopTechnologies(ctx, 10)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id proyek tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NotFound("proyek tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.NewError(fiber.StatusConflict, "PROJECT_LIMIT", "jumlah proyek sudah mencapai batas")
	default:
		return helper.Internal(err)
	}
}

func NormalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	seen := make(map[string]struct{})
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		cleaned := strings.ToLower(strings.TrimSpace(tag))
		if cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; !ok {
			seen[cleaned] = struct{}{}
			result = append(result, cleaned)
		}
	}
	return result
}