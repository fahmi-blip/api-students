package main
import (
 	"sort"
 	"strconv"
 	"strings"
 	
	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
	"api-students/app/repository"
)

var students []Student
var nextID = 1

func findUserIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func cocokPencarian(u Student, kata string) bool {
	kata = strings.ToLower(kata)
	return strings.Contains(strings.ToLower(u.Name), kata) ||
		strings.Contains(strings.ToUpper(u.Name), kata)
}

func paramID(c *fiber.Ctx) (int, bool){
	id, err :=strconv.Atoi(c.Params("id"))
	if err != nil || id <1 {
		return 0, false
	}
	return id, true
}

func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	//1. Saring
	hasil := []Student{}
	for _, u := range students {
		if q.IsActive != nil && u.IsActive != *q.IsActive {
			continue
		}
		if q.MinGrade != nil && u.Grade < *q.MinGrade{
			continue
		}
		if q.MaxGrade != nil && u.Grade < *q.MaxGrade{
			continue
		}
		if q.Search != "" && !cocokPencarian(u, q.Search) {
			continue
		}
		hasil = append(hasil, u)
	}

	//2. Urutkan
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim":
			lebihKecil = hasil[i].Nim < hasil[j].Nim
		case "name":
			lebihKecil = hasil[i].Name < hasil[j].Name
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		default:
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	//3. Potong sesuai halaman
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai >total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar user berhasil diambil", hasil[mulai:akhir], &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPage: totalPages,
	})
}

func getUser(c *fiber.Ctx) error {
 	id, valid := paramID(c)
 	if !valid {
 		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
 	}
 	i := findUserIndex(id)
 	if i == -1 {
 		return fail(c, fiber.StatusNotFound, "user tidak ditemukan")
 	}
 	return ok(c, "user ditemukan", students[i])
}

func createUser(c *fiber.Ctx) error {
 	var req CreatedUserRequest
 	if err := c.BodyParser(&req); err != nil {
 		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
 	}
 	errs := map[string]string{}
 	req.Nim = strings.TrimSpace(req.Nim)
 	req.Name = strings.TrimSpace(req.Name)
 	
	if req.Nim == "" {
 		errs["nim"] = "wajib diisi"
 	}
 	if req.Name == "" {
 		errs["name"] = "format nama tidak valid"
 	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus di antara 0 dan 100"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}


 	for _, u := range students {
 		if u.Nim == req.Nim {
 			return fail(c, fiber.StatusConflict, "Nim sudah dipakai")
 		}
 	}

 	baru := Student{
 		ID: nextID,
 		Nim: req.Nim,
 		Name: req.Name,
 		Grade: req.Grade,
 		IsActive: true,
 	}
 	students = append(students, baru)
 	nextID++
	
 	return created(c, "user berhasil dibuat", baru,
		"/api/v1/students/"+strconv.Itoa(baru.ID))
}

// PUT mengganti SELURUH isi. Field yang tidak dikirim dianggap dikosongkan,
// karena itu semuanya wajib ada.
func replaceUser(c *fiber.Ctx) error {
 	id, valid := paramID(c)
 	if !valid {
 		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
 	}
 	i := findUserIndex(id)
 	if i == -1 {
 		return fail(c, fiber.StatusNotFound, "user tidak ditemukan")
 	}
 	var req ReplaceUserRequest
 	if err := c.BodyParser(&req); err != nil {
 		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
 	}
 	errs := map[string]string{}
 	if strings.TrimSpace(req.Name) == "" {
 		errs["name"] = "wajib diisi pada PUT"
 	}
 	if req.Grade < 0 || req.Grade > 100 {
 		errs["grade"] = "harus di antara 0 dan 100"
 	}
 	if len(errs) > 0 {
 		return failValidation(c, errs)
 	}
 	students[i].Name = req.Name
 	students[i].Grade = req.Grade
 	students[i].IsActive = req.IsActive
 	
	return ok(c, "user berhasil diganti seluruhnya", students[i])
}

// PATCH hanya mengubah field yang benar-benar dikirim.
func patchUser(c *fiber.Ctx) error {
 	id, valid := paramID(c)
 	if !valid {
 		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
 	}
 	i := findUserIndex(id)
 	if i == -1 {
 		return fail(c, fiber.StatusNotFound, "user tidak ditemukan")
 	}
 	var req PatchUserRequest
 	if err := c.BodyParser(&req); err != nil {
 		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
 	}
 	if req.Name == nil && req.Grade == nil && req.IsActive == nil {
 		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
 	}
 	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
 			return failValidation(c, map[string]string{"Name": "tidak boleh kosong"})
 		}
 		students[i].Name = *req.Name
 	}
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			return failValidation(c, map[string]string{"grade": "harus di antara 0 dan 100"})
		}
		students[i].Grade = *req.Grade
	}
 	if req.IsActive != nil {
 		students[i].IsActive = *req.IsActive
 	}
 	return ok(c, "user berhasil diperbarui sebagian", students[i])
}

func deleteUser(c *fiber.Ctx) error {
 	id, valid := paramID(c)
 	if !valid {
 		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
 	}
 	i := findUserIndex(id)
 	if i == -1 {
 		return fail(c, fiber.StatusNotFound, "user tidak ditemukan")
 	}
 	students = append(students[:i], students[i+1:]...)
 	
	return noContent(c) // 204: berhasil, dan memang tidak ada yang perlu dikirim
}
