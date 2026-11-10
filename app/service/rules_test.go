package service

import "testing"

func TestMaxSKS(t *testing.T) {
	cases := []struct {
		name    string
		ipk     float64
		wantSKS int
	}{
		{"IPK tinggi (3.85)", 3.85, 24},
		{"Batas IPK 3.00", 3.00, 24},
		{"Batas atas IPK 2.50-2.99 (2.99)", 2.99, 21},
		{"IPK menengah (2.75)", 2.75, 21},
		{"Batas bawah IPK 2.50", 2.50, 21},
		{"Batas atas IPK < 2.50 (2.49)", 2.49, 18},
		{"IPK rendah (1.50)", 1.50, 18},
		{"IPK nol (0.00)", 0.00, 18},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MaxSKS(tc.ipk)
			if got != tc.wantSKS {
				t.Errorf("MaxSKS(%v) = %d, ingin %d", tc.ipk, got, tc.wantSKS)
			}
		})
	}
}

func TestValidateTahunAkademik(t *testing.T) {
	if err := ValidateTahunAkademik("2026/2027 Ganjil"); err != nil {
		t.Errorf("seharusnya valid, tetapi dapat error: %v", err)
	}
	if err := ValidateTahunAkademik(""); err == nil {
		t.Error("string kosong seharusnya tidak valid")
	}
	if err := ValidateTahunAkademik("   "); err == nil {
		t.Error("hanya spasi seharusnya tidak valid")
	}
}

