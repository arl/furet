package main_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// goTool reports the path to the Go tool.
func goTool(t *testing.T) string {
	var exeSuffix string
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	path := filepath.Join(runtime.GOROOT(), "bin", "go"+exeSuffix)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	goBin, err := exec.LookPath("go" + exeSuffix)
	if err != nil {
		t.Fatalf("cannot find go tool: " + err.Error())
	}
	return goBin
}

// furetPath builds furet in a temporary directory and reports its path.
func furetPath(t *testing.T) string {
	tmpDir := t.TempDir()
	exepath := filepath.Join(tmpDir, "furet.exe")

	out, err := exec.Command(goTool(t), "build", "-o", exepath, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build -o %v . : %v\n%s", exepath, err, string(out))
	}
	return exepath
}

func generateKey(t *testing.T, furetPath string) string {
	buf, err := exec.Command(furetPath, "-g").Output()
	if err != nil {
		t.Fatalf("furet -g: %v", err)
	}

	if len(buf) == 0 {
		t.Errorf("empty key")
	}
	return string(buf)
}

func TestFuretGenerateKey(t *testing.T) {
	furetPath := furetPath(t)
	generateKey(t, furetPath)
}

func TestFuretEncryptDecryptFile(t *testing.T) {
	tmpDir := t.TempDir()
	furetPath := furetPath(t)
	key := generateKey(t, furetPath)

	original := filepath.Join("testdata", "Isaac.Newton-Opticks.txt")
	encrypted := filepath.Join(tmpDir, "txt.furet")

	if _, err := exec.Command(furetPath, "--key", key, "--output", encrypted, original).Output(); err != nil {
		t.Fatalf("furet --key %s --output %s %s: %v", key, encrypted, original, err)
	}
	if _, err := os.Stat(encrypted); os.IsNotExist(err) {
		t.Fatalf("encrypted file not found: %s", err)
	}

	decoded := filepath.Join(tmpDir, "decoded")
	if _, err := exec.Command(furetPath, "--decrypt", "--key", key, "--output", decoded, encrypted).Output(); err != nil {
		t.Fatalf("furet --decrypt --key %s --output %s %s: %v", key, decoded, encrypted, err)
	}
	if _, err := os.Stat(decoded); os.IsNotExist(err) {
		t.Fatalf("decoded file not found: %s", err)
	}

	var (
		dst, org []byte
		err      error
	)
	if dst, err = os.ReadFile(decoded); err != nil {
		t.Fatal(err)
	}
	if org, err = os.ReadFile(original); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(dst, org) {
		t.Errorf("decrypted file differs from original")
	}
}

func execPipe(t *testing.T, input []byte, name string, arg ...string) []byte {
	cmd := exec.Command(name, arg...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, bytes.NewReader(input))
	}()

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatal(err)
	}

	cmderr, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}

	if len(cmderr) != 0 {
		t.Fatalf("stderr: %v", string(cmderr))
	}

	return out
}

func TestFuretEncryptDecryptPipe(t *testing.T) {
	furetPath := furetPath(t)
	key := generateKey(t, furetPath)

	buforg, err := os.ReadFile(filepath.Join("testdata", "Isaac.Newton-Opticks.txt"))
	if err != nil {
		t.Fatal(err)
	}

	crypted := execPipe(t, buforg, furetPath, "--key", key)
	decrypted := execPipe(t, crypted, furetPath, "--decrypt", "--key", key)

	if !bytes.Equal(buforg, decrypted) {
		t.Errorf("decrypted buffers differs from original")
	}
}
