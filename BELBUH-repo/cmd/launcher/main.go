//go:build windows

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const avTunDownloadURL = "https://avtunproxy.by/windows/"

var (
	shell32      = syscall.NewLazyDLL("shell32.dll")
	shellExecute = shell32.NewProc("ShellExecuteW")
	user32       = syscall.NewLazyDLL("user32.dll")
	messageBox   = user32.NewProc("MessageBoxW")
)

func utf16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func main() {
	exe, err := os.Executable()
	if err != nil {
		showError("Не удалось определить папку приложения.")
		return
	}
	exe, _ = filepath.Abs(exe)
	registerProtocol(exe)
	if len(os.Args) > 1 {
		raw := strings.TrimRight(os.Args[1], "/")
		if strings.HasPrefix(strings.ToLower(raw), "belhub://") {
			handleCommand(raw)
			return
		}
	}
	page := filepath.Join(filepath.Dir(exe), "BELHUB-3.2.html")
	if _, err := os.Stat(page); err != nil {
		showError("Файл BELHUB-3.2.html не найден рядом с программой запуска.\r\n\r\nРаспакуйте архив полностью в одну папку.")
		return
	}
	if !open(page, filepath.Dir(exe)) {
		showError("Не удалось открыть BELHUB-3.2.html в браузере.")
	}
}

func registerProtocol(exe string) {
	key := `HKCU\Software\Classes\belhub`
	command := `"` + exe + `" "%1"`
	commands := [][]string{
		{"add", key, "/ve", "/d", "URL:BELHUB Local Protocol", "/f"},
		{"add", key, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", key + `\DefaultIcon`, "/ve", "/d", exe + ",0", "/f"},
		{"add", key + `\shell\open\command`, "/ve", "/d", command, "/f"},
	}
	for _, args := range commands {
		_ = exec.Command("reg.exe", args...).Run()
	}
}

func handleCommand(raw string) {
	lower := strings.ToLower(strings.TrimRight(raw, "/"))
	if strings.HasPrefix(lower, "belhub://avtun") {
		openWithAvTun(raw)
		return
	}
	switch lower {
	case "belhub://certificates":
		openAvestCertificates()
	case "belhub://support":
		requestSupport()
	default:
		showError("Неизвестная команда BELHUB.")
	}
}

func requestSupport() {
	if app := findAnyDesk(); app != "" {
		if !open(app, filepath.Dir(app)) {
			showError("AnyDesk найден, но его не удалось запустить.")
		}
		return
	}
	open("https://anydesk.com/ru/downloads/thank-you?dv=win_exe", "")
}

func findAnyDesk() string {
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("APPDATA"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("USERPROFILE"), "Downloads", "AnyDesk.exe"),
		filepath.Join(os.Getenv("USERPROFILE"), "Desktop", "AnyDesk.exe"),
	}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	if path, err := exec.LookPath("AnyDesk.exe"); err == nil {
		return path
	}
	return ""
}

func openAvestCertificates() {
	update := findAvestFile([]string{"get_crl.bat", "getcrl.bat", "updatecrl.exe", "avcrlupdate.exe"}, true)
	manager := findAvestFile([]string{"avpcm.exe", "avpcmex.exe", "avpcm_m.exe"}, false)
	if update != "" {
		if strings.HasSuffix(strings.ToLower(update), ".bat") {
			_ = exec.Command("cmd.exe", "/c", "start", "", update).Start()
		} else {
			open(update, filepath.Dir(update))
		}
	}
	if manager != "" {
		open(manager, filepath.Dir(manager))
	}
	if update == "" && manager == "" {
		showInfo("Компоненты АВЕСТ не найдены.\r\n\r\nУстановите подходящий комплект АВЕСТ через меню ПО, затем повторите действие.")
		return
	}
	msg := ""
	if update != "" {
		msg += "Запущено обновление сертификатов и списков отозванных сертификатов (СОС).\r\n"
	} else {
		msg += "Автоматическая утилита обновления СОС не найдена. В Персональном менеджере выберите:\r\nСервис → Обновление СОС и сертификатов УЦ.\r\n"
	}
	if manager != "" {
		msg += "\r\nОткрыт Персональный менеджер АВЕСТ. В разделе «Личные» отображаются установленные сертификаты и сроки их действия."
	} else {
		msg += "\r\nПерсональный менеджер АВЕСТ не найден."
	}
	showInfo(msg)
}

func findAvestFile(names []string, includeShortcuts bool) string {
	wanted := map[string]bool{}
	for _, n := range names {
		wanted[strings.ToLower(n)] = true
	}
	roots := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Avest"), filepath.Join(os.Getenv("ProgramFiles(x86)"), "Avest"),
		filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(os.Getenv("PUBLIC"), "Desktop"), filepath.Join(os.Getenv("USERPROFILE"), "Desktop"),
	}
	for _, root := range roots {
		if root == "" {
			continue
		}
		var found string
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				rel, _ := filepath.Rel(root, path)
				if strings.Count(rel, string(os.PathSeparator)) > 5 {
					return filepath.SkipDir
				}
				return nil
			}
			name := strings.ToLower(d.Name())
			if wanted[name] {
				found = path
				return filepath.SkipAll
			}
			if includeShortcuts && strings.HasSuffix(name, ".lnk") && (strings.Contains(name, "сос") || strings.Contains(name, "crl")) {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			return found
		}
	}
	return ""
}

func openWithAvTun(raw string) {
	target := avTunTarget(raw)
	if target == "" {
		showError("Не указан адрес портала для проверки AvTunProxy.")
		return
	}
	if ensureAvTunProxy() {
		if !open(target, "") {
			showError("AvTunProxy найден, но портал открыть не удалось.")
		}
		return
	}
	if askYesNo("Для выбранного портала требуется установленный и запущенный AvTunProxy.\r\n\r\nAvTunProxy не найден на этом компьютере. Скачать установщик с официальной страницы ЗАО «АВЕСТ»?") {
		open(avTunDownloadURL, "")
	}
}

func avTunTarget(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	q := parsed.Query().Get("target")
	if q == "" {
		return ""
	}
	decoded, err := url.QueryUnescape(q)
	if err == nil {
		q = decoded
	}
	u, err := url.Parse(q)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || !isAvTunPortalHost(u.Hostname()) {
		return ""
	}
	return q
}

func isAvTunPortalHost(host string) bool {
	switch strings.ToLower(host) {
	case "portal2.ssf.gov.by", "portal.ssf.gov.by", "rvd.nbrb.by":
		return true
	default:
		return false
	}
}

func ensureAvTunProxy() bool {
	if isAvTunProxyRunning() {
		return true
	}
	if app := findAvTunProxy(); app != "" {
		_ = open(app, filepath.Dir(app))
		deadline := time.Now().Add(4 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(500 * time.Millisecond)
			if isAvTunProxyRunning() {
				return true
			}
		}
		// AvTunProxy может быть установлен, но запускать локальный PAC только после
		// ручного подтверждения пользователя. В этом случае всё равно открываем портал.
		return true
	}
	return false
}

func isAvTunProxyRunning() bool {
	client := http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:10224/proxy.pac")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}

func findAvTunProxy() string {
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "AvTunProxy", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "AvTunProxy", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Avest", "AvTunProxy", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Avest", "AvTunProxy", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Avest", "AvTun", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Avest", "AvTun", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "AvTunProxy", "AvTunProxy.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "AvTunProxy", "AvTunProxy.exe"),
	}
	for _, path := range candidates {
		if fileExists(path) {
			return path
		}
	}
	if path, err := exec.LookPath("AvTunProxy.exe"); err == nil && fileExists(path) {
		return path
	}
	roots := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Avest"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Avest"),
		filepath.Join(os.Getenv("ProgramFiles"), "AvTunProxy"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "AvTunProxy"),
	}
	for _, root := range roots {
		if root == "" || !dirExists(root) {
			continue
		}
		var found string
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				rel, _ := filepath.Rel(root, path)
				if strings.Count(rel, string(os.PathSeparator)) > 4 {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.EqualFold(d.Name(), "AvTunProxy.exe") {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			return found
		}
	}
	return ""
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func askYesNo(text string) bool {
	const mbYesNoIconQuestion = 0x00000004 | 0x00000020
	r, _, _ := messageBox.Call(0, uintptr(unsafe.Pointer(utf16(text))), uintptr(unsafe.Pointer(utf16("BELHUB 3.2 — AvTunProxy"))), mbYesNoIconQuestion)
	return r == 6
}

func open(target, dir string) bool {
	r, _, _ := shellExecute.Call(0, uintptr(unsafe.Pointer(utf16("open"))), uintptr(unsafe.Pointer(utf16(target))), 0, uintptr(unsafe.Pointer(utf16(dir))), 1)
	return r > 32
}
func showError(text string) {
	messageBox.Call(0, uintptr(unsafe.Pointer(utf16(text))), uintptr(unsafe.Pointer(utf16("BELHUB 3.2"))), 0x10)
}
func showInfo(text string) {
	showInfoTitle("BELHUB 3.2 — сертификаты и СОС", text)
}
func showInfoTitle(title, text string) {
	messageBox.Call(0, uintptr(unsafe.Pointer(utf16(text))), uintptr(unsafe.Pointer(utf16(title))), 0x40)
}

var _ = fmt.Sprintf
