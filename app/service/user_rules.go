package service

import (
	"strings"

	"api-students/app/model"
)

// File ini berisi business rules MURNI: tidak menyentuh fiber.Ctx, 
// tidak menyentuh database, dan tidak tahu apa pun tentang HTTP. 

// ValidateCreate memeriksa isi permintaan pembuatan user.  
// Mengembalikan peta berisi field yang bermasalah; kosong berarti lolos. 
func ValidateCreate(req model.CreatedStudentRequest) map[string]string { 
	errs := map[string]string{} 
	
	if strings.TrimSpace(req.Nim) == "" { 
		errs["nim"] = "nim wajib diisi" 
	} 
	if strings.TrimSpace(req.Name) == "" { 
		errs["name"] = "name wajib diisi" 
	} 
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus di antara 0 dan 100"
	}
	return errs 
} 

// ValidateReplace memeriksa isi permintaan PUT. 
// Seluruh field wajib ada karena PUT mengganti isi secara keseluruhan. 
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string { 
	errs := map[string]string{} 
	
	if strings.TrimSpace(
		req.Name) == "" { errs["username"] = "wajib diisi pada PUT" 
	} 
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "wajib diisi dan nilai di antara 0 & 100 pada PUT"
	}
	return errs
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada. 
// Field yang bernilai nil dibiarkan apa adanya. 
 
func ApplyPatch( 
	current model.Student, req model.PatchStudentRequest, 
) (model.Student, map[string]string) { 
	errs := map[string]string{} 
	
	if req.Name != nil { 
		if strings.TrimSpace(*req.Name) == "" { 
			errs["username"] = "tidak boleh kosong" 
		} else { 
			current.Name = *req.Name 
		} 
	} 
	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "format grade tidak valid (harus pada rentang 0-100)"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil { 
		current.IsActive = *req.IsActive 
	} 
	return current, errs 
} 
	
// IsEmptyPatch menandai permintaan PATCH yang tidak mengubah apa pun. 
func IsEmptyPatch(req model.PatchStudentRequest) bool { 
	return req.Name == nil && req.Grade == nil && req.IsActive == nil 
} 

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan. 
func CountTotalPages(total, limit int) int { 
	if limit <= 0 { 
		return 0 
	} 
	return (total + limit - 1) / limit 
} 
