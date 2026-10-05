//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

var shell = syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")

func w(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func main() {
	e, _ := os.Executable()
	d := filepath.Dir(e)
	registerMainProtocol(filepath.Join(d, "BELHUB-3.2.exe"))
	page := filepath.Join(d, "widget.html")
	uri := "file:///" + filepath.ToSlash(page)
	paths := []string{filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"), filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe")}
	for _, p := range paths {
		if _, x := os.Stat(p); x == nil {
			_ = exec.Command(p, "--app="+uri, "--window-size=410,590").Start()
			return
		}
	}
	shell.Call(0, uintptr(unsafe.Pointer(w("open"))), uintptr(unsafe.Pointer(w(page))), 0, uintptr(unsafe.Pointer(w(d))), 1)
}

func registerMainProtocol(mainExe string) {
	if info, err := os.Stat(mainExe); err != nil || info.IsDir() {
		return
	}
	key := `HKCU\Software\Classes\belhub`
	command := `"` + mainExe + `" "%1"`
	commands := [][]string{
		{"add", key, "/ve", "/d", "URL:BELHUB Local Protocol", "/f"},
		{"add", key, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", key + `\DefaultIcon`, "/ve", "/d", mainExe + ",0", "/f"},
		{"add", key + `\shell\open\command`, "/ve", "/d", command, "/f"},
	}
	for _, args := range commands {
		_ = exec.Command("reg.exe", args...).Run()
	}
}
