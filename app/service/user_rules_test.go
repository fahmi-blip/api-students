package service

import (
	"testing" 
	"api-students/app/model"
) 

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh 
// database, dan tidak membuat fiber.Ctx. 
 
func TestCountTotalPages(t *testing.T) { 
	cases := []struct{ total, limit, want int }{ 
		{0, 10, 0}, 
		{1, 10, 1}, 
		{10, 10, 1}, 
		{11, 10, 2}, 
		{137, 20, 7}, 
	} 
	
	for _, tc := range cases { 
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want { 
			t.Errorf("total=%d limit=%d: harap %d, dapat %d", 
				tc.total, tc.limit, tc.want, got) 
		} 
	} 
} 
func TestApplyPatch(t *testing.T) { 
	initial := model.Student{ID: 1, Nim: "434241120", Name: "sari", Grade:84, IsActive: true} 
	inactive := false 
	
	result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive}) 
	
	if len(errs) != 0 { 
		t.Fatalf("tidak seharusnya ada error: %v", errs) 
	} 
	if result.Name != "sari" { 
		t.Error("field yang tidak dikirim seharusnya tidak berubah") 
	} 
	if result.Grade != initial.Grade {
			t.Error("grade seharusnya tidak berubah ketika nilainya invalid")
	}
	if result.IsActive { 
		t.Error("is_active seharusnya berubah menjadi false") 
	} 
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("patch tanpa field seharusnya dianggap kosong")
	}
 
	nama := "Budi"
	if IsEmptyPatch(model.PatchStudentRequest{Name: &nama}) {
		t.Error("patch dengan satu field terisi seharusnya tidak dianggap kosong")
	}
}