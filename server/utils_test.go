package server

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestCheckError(t *testing.T) {
	err := errors.New("TestError")
	const str = "Test string"
	a := CheckError(nil, str, false)
	if !a {
		t.Errorf("CheckError(%v, '%v', %v) - return %v, wait %v", nil, str, false, a, true)
	}
	a = CheckError(nil, str, true)
	if !a {
		t.Errorf("CheckError(%v, '%v', %v) - return %v, wait %v", nil, str, true, a, true)
	}
	a = CheckError(err, str, false)
	if a {
		t.Errorf("CheckError(%v, '%v', %v) - return %v, wait %v", err, str, false, a, false)
	}
}

func TestCheckErrorCrasher(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		CheckError(errors.New("New error"), "Test string", true)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckErrorCrasher")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestGetMD5Hash(t *testing.T) {
	var str string
	str = GetMD5Hash("")
	if "d41d8cd98f00b204e9800998ecf8427e" != str {
		t.Errorf("GetMD5Hash('%v') = %v, wait %v", "", str, "d41d8cd98f00b204e9800998ecf8427e")
	}
	str = GetMD5Hash("19643")
	if "18d596dcf73043e0c8a6e3bfef2a0731" != str {
		t.Errorf("GetMD5Hash('%v') = %v, wait %v", "19643", str, "18d596dcf73043e0c8a6e3bfef2a0731")
	}
	str = GetMD5Hash("Test string")
	if "0fd3dbec9730101bff92acc820befc34" != str {
		t.Errorf("GetMD5Hash('%v') = %v, wait %v", "Test string", str, "0fd3dbec9730101bff92acc820befc34")
	}
	str = GetMD5Hash("Русский тестовый текст")
	if "adba8e9ba3a55ebe7fad308b33d04001" != str {
		t.Errorf("GetMD5Hash('%v') = %v, wait %v", "Русский тестовый текст", str, "adba8e9ba3a55ebe7fad308b33d04001")
	}
}
