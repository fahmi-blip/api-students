package service

import (
	"errors"
	"strings"
)

// MaxSKS menentukan batas SKS per semester berdasarkan IPK terakhir:
// 1. IPK terakhir:
//    a. IPK >= 3.00: maksimal 24 SKS
//    b. IPK 2.50 - 2.99: maksimal 21 SKS
//    c. IPK < 2.50: maksimal 18 SKS
func MaxSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

// ValidateTahunAkademik memvalidasi nilai tahun akademik
func ValidateTahunAkademik(ta string) error {
	ta = strings.TrimSpace(ta)
	if ta == "" {
		return errors.New("tahun akademik wajib diisi")
	}
	return nil
}

