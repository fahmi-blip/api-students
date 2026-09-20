package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh menyentuh satu baris
// data student. Dua jalur yang diizinkan:
//  1. Kepemilikan — data itu didaftarkan olehnya sendiri (ownerID == UserID).
//  2. Permission — role-nya memang berhak atas data siapa pun.
//
// ownerID bernilai 0 untuk data lama yang belum punya pemilik; karena tidak
// ada user ber-id 0, jalur kepemilikan otomatis gagal untuk baris semacam
// itu dan keputusan sepenuhnya bergantung pada permission :any.
//
// File ini murni: tidak mengimpor fiber maupun repository, sehingga bisa
// diuji hanya dengan memanggilnya langsung.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if ownerID != 0 && current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}