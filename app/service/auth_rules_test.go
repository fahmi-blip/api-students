package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateRegister_PasswordTerlaluPendek(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "abc123",
	}

	errs := ValidateRegister(req)

	if _, ada := errs["password"]; !ada {
		t.Errorf("diharapkan ada error pada field password, tapi tidak ada. errs=%v", errs)
	}
}

func TestValidateRegister_PasswordTanpaAngka(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "hanyahuruf",
	}

	errs := ValidateRegister(req)

	if _, ada := errs["password"]; !ada {
		t.Errorf("diharapkan ditolak karena tidak memuat angka, errs=%v", errs)
	}
}

func TestValidateRegister_PasswordUmumDitolak(t *testing.T) {
	req := model.RegisterRequest{
		Username: "budi",
		Email:    "budi@example.com",
		Password: "password1",
	}

	errs := ValidateRegister(req)

	if msg, ada := errs["password"]; !ada || msg != "password terlalu umum" {
		t.Errorf("diharapkan pesan 'password terlalu umum', dapat: %v", errs)
	}
}

func TestValidateRegister_UsernameTerlaluPendek(t *testing.T) {
	req := model.RegisterRequest{
		Username: "ab",
		Email:    "ab@example.com",
		Password: "rahasia123",
	}

	errs := ValidateRegister(req)

	if _, ada := errs["username"]; !ada {
		t.Errorf("diharapkan ada error pada field username, errs=%v", errs)
	}
}

func TestValidateRegister_EmailTidakValid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "bukan-email",
		Password: "rahasia123",
	}

	errs := ValidateRegister(req)

	if _, ada := errs["email"]; !ada {
		t.Errorf("diharapkan ada error pada field email, errs=%v", errs)
	}
}

func TestValidateRegister_InputValidTidakAdaError(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "rahasia123",
	}

	errs := ValidateRegister(req)

	if len(errs) != 0 {
		t.Errorf("diharapkan tidak ada error untuk input valid, dapat: %v", errs)
	}
}

func TestValidateLogin_FieldKosong(t *testing.T) {
	req := model.LoginRequest{Username: "", Password: ""}

	errs := ValidateLogin(req)

	if len(errs) != 2 {
		t.Errorf("diharapkan 2 error (username dan password), dapat: %v", errs)
	}
}
