//go:build windows

package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type Portal struct {
	Name, Category, URL, Note string
	Legacy                    bool
}

var portals = []Portal{
	{"МНС — официальный сайт", "Налоги и финансы", "https://nalog.gov.by/", "Новости, сервисы и нормативные материалы", false},
	{"МНС — личный кабинет", "Налоги и финансы", "https://portal.nalog.gov.by/", "Вход по ЭЦП или учетной записи", true},
	{"МНС — личный кабинет физлица", "Налоги и финансы", "https://lkfl.portal.nalog.gov.by/", "Сервисы для физических лиц", false},
	{"Министерство финансов", "Налоги и финансы", "https://www.minfin.gov.by/", "Официальный сайт Минфина", false},
	{"Национальный банк", "Налоги и финансы", "https://www.nbrb.by/", "Курсы, статистика и нормативные документы", false},
	{"ФСЗН — корпоративный портал", "Отчетность", "https://portal2.ssf.gov.by/mainPage/?step=payer", "Электронные документы ФСЗН", true},
	{"ФСЗН — официальный сайт", "Отчетность", "https://ssf.gov.by/", "Фонд социальной защиты населения", false},
	{"Белстат — электронный респондент", "Отчетность", "https://e-respondent.belstat.gov.by/", "Электронная статистическая отчетность", true},
	{"Белстат", "Отчетность", "https://www.belstat.gov.by/", "Официальная статистика", false},
	{"Белгосстрах", "Отчетность", "https://bgs.by/", "Официальный сайт", false},
	{"Белгосстрах — кабинет страхователя", "Отчетность", "https://lk.bgs.by/", "Электронные сервисы страхователя", false},
	{"ЕПЭУ — электронные услуги", "Госуслуги", "https://e-pasluga.by/", "Единый портал электронных услуг", false},
	{"Личный кабинет ЕПЭУ", "Госуслуги", "https://account.gov.by/", "Единый личный кабинет", false},
	{"Электронные обращения", "Госуслуги", "https://обращения.бел/", "Обращения граждан и юридических лиц", false},
	{"НЦЭУ", "ЭЦП и доступ", "https://nces.by/", "Национальный центр электронных услуг", false},
	{"НЦЭУ — техподдержка", "ЭЦП и доступ", "https://helpdesk.nces.by/", "Инструкции по АВЕСТ и ГосСУОК", false},
	{"НЦЭУ — облачное хранилище", "ЭЦП и доступ", "https://store.nces.by/", "Комплект абонента и сертификаты", false},
	{"АВЕСТ — загрузки", "ЭЦП и доступ", "https://www.avest.by/crypto/download/", "Официальные версии криптографического ПО", false},
	{"ГосСУОК", "ЭЦП и доступ", "https://pki.gov.by/", "Инфраструктура открытых ключей", true},
	{"ЕГР", "Бизнес", "https://egr.gov.by/", "Единый государственный регистр юрлиц и ИП", false},
	{"Единый реестр лицензий", "Бизнес", "https://license.gov.by/", "Сведения о лицензиях", false},
	{"Портал госзакупок", "Закупки", "https://goszakupki.by/", "Электронные государственные закупки", false},
	{"БУТБ", "Закупки", "https://butb.by/", "Белорусская универсальная товарная биржа", false},
	{"Национальный правовой портал", "Право", "https://pravo.by/", "Законодательство Республики Беларусь", false},
	{"Судебные кабинеты", "Право", "https://court.gov.by/", "Система общих судов", false},
	{"Министерство юстиции", "Право", "https://minjust.gov.by/", "Официальный сайт", false},
	{"Электронная таможня", "Таможня", "https://www.customs.gov.by/electronic-customs/", "Электронные сервисы таможни", false},
	{"Таможенный портал", "Таможня", "https://gtk.gov.by/", "Государственный таможенный комитет", false},
	{"Министерство труда", "Органы власти", "https://mintrud.gov.by/", "Официальный сайт", false},
	{"Министерство экономики", "Органы власти", "https://economy.gov.by/", "Официальный сайт", false},
	{"Министерство связи", "Органы власти", "https://mpt.gov.by/", "Официальный сайт", false},
	{"Министерство образования", "Органы власти", "https://edu.gov.by/", "Официальный сайт", false},
	{"Министерство здравоохранения", "Органы власти", "https://minzdrav.gov.by/", "Официальный сайт", false},
	{"Мингорисполком", "Органы власти", "https://minsk.gov.by/", "Официальный портал Минска", false},
	{"Совет Министров", "Органы власти", "https://government.by/", "Правительство Республики Беларусь", false},
	{"Президент Республики Беларусь", "Органы власти", "https://president.gov.by/", "Официальный интернет-портал", false},
}

var mw *walk.MainWindow
var search *walk.LineEdit
var category *walk.ComboBox
var list *walk.ListBox
var details *walk.TextEdit
var logBox *walk.TextEdit
var status *walk.StatusBarItem
var visible []Portal

func categories() []string {
	m := map[string]bool{}
	for _, p := range portals {
		m[p.Category] = true
	}
	r := []string{"Все категории"}
	for c := range m {
		r = append(r, c)
	}
	sort.Strings(r[1:])
	return r
}

func refresh() {
	if list == nil || search == nil || category == nil {
		return
	}
	q := strings.ToLower(strings.TrimSpace(search.Text()))
	cat := category.Text()
	visible = nil
	names := []string{}
	for _, p := range portals {
		if (cat == "" || cat == "Все категории" || p.Category == cat) &&
			(q == "" || strings.Contains(strings.ToLower(p.Name+" "+p.Note+" "+p.URL), q)) {
			visible = append(visible, p)
			name := p.Name
			if p.Legacy {
				name += "   • ЭЦП"
			}
			names = append(names, name)
		}
	}
	list.SetModel(names)
	status.SetText(fmt.Sprintf("Порталов: %d   |   BELHUB не хранит пароли и закрытые ключи", len(visible)))
	if len(visible) > 0 {
		list.SetCurrentIndex(0)
		showDetails()
	} else {
		details.SetText("Ничего не найдено.")
	}
}

func selected() (Portal, bool) {
	i := list.CurrentIndex()
	if i < 0 || i >= len(visible) {
		return Portal{}, false
	}
	return visible[i], true
}

func showDetails() {
	p, ok := selected()
	if !ok {
		return
	}
	req := "Открывается в современном браузере."
	if p.Legacy {
		req = "Для входа может потребоваться АВЕСТ, AvCMXWebP или режим Internet Explorer в Microsoft Edge."
	}
	details.SetText(p.Name + "\r\n\r\n" + p.Note + "\r\n\r\n" + req + "\r\n\r\nАдрес:\r\n" + p.URL)
}

func reportError(action string, err error) {
	line := time.Now().Format("15:04:05") + "  ОШИБКА: " + action + ": " + err.Error()
	logBox.AppendText(line + "\r\n")
	walk.MsgBox(mw, "BELHUB — ошибка", line, walk.MsgBoxIconError)
}

func addLog(s string) { logBox.AppendText(time.Now().Format("15:04:05") + "  " + s + "\r\n") }

func openDefault(rawURL string) {
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", rawURL)
	if err := cmd.Start(); err != nil {
		reportError("открытие ссылки", err)
		return
	}
	addLog("Открыта ссылка: " + rawURL)
}

func openEdge(rawURL string) {
	paths := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err = exec.Command(path, "--new-window", rawURL).Start(); err != nil {
				reportError("запуск Edge", err)
			} else {
				addLog("Портал открыт в Edge")
			}
			return
		}
	}
	openDefault(rawURL)
}

func addTrusted() {
	p, ok := selected()
	if !ok {
		return
	}
	u, err := url.Parse(p.URL)
	if err != nil {
		reportError("разбор адреса", err)
		return
	}
	host := strings.TrimPrefix(u.Hostname(), "www.")
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		reportError("домен", fmt.Errorf("некорректный адрес"))
		return
	}
	root := strings.Join(parts[len(parts)-2:], ".")
	sub := strings.Join(parts[:len(parts)-2], ".")
	key := `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings\ZoneMap\Domains\` + root
	if sub != "" {
		key += `\` + sub
	}
	if walk.MsgBox(mw, "Надежные сайты", "Добавить "+host+" в зону «Надежные сайты» текущего пользователя?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
		return
	}
	cmd := exec.Command("reg.exe", "add", key, "/v", "https", "/t", "REG_DWORD", "/d", "2", "/f")
	if out, err := cmd.CombinedOutput(); err != nil {
		reportError("изменение зоны безопасности", fmt.Errorf("%v: %s", err, out))
		return
	}
	addLog(host + " добавлен в надежные сайты")
	walk.MsgBox(mw, "BELHUB", "Настройка применена для "+host, walk.MsgBoxIconInformation)
}

func diagnostics() {
	addLog("Запущена диагностика…")
	go func() {
		lines := []string{"Диагностика рабочего места:"}
		edge := false
		for _, p := range []string{filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"), filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe")} {
			if _, e := os.Stat(p); e == nil {
				edge = true
			}
		}
		if edge {
			lines = append(lines, "✓ Microsoft Edge найден")
		} else {
			lines = append(lines, "✗ Microsoft Edge не найден")
		}
		avest := false
		for _, p := range []string{filepath.Join(os.Getenv("ProgramFiles"), "Avest"), filepath.Join(os.Getenv("ProgramFiles(x86)"), "Avest")} {
			if s, e := os.Stat(p); e == nil && s.IsDir() {
				avest = true
			}
		}
		if avest {
			lines = append(lines, "✓ Каталог АВЕСТ найден")
		} else {
			lines = append(lines, "! АВЕСТ не найден по стандартному пути")
		}
		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Get("https://nces.by/")
		if err == nil {
			resp.Body.Close()
			lines = append(lines, "✓ HTTPS-доступ к НЦЭУ работает")
		} else {
			lines = append(lines, "✗ HTTPS-доступ к НЦЭУ: "+err.Error())
		}
		mw.Synchronize(func() {
			for _, x := range lines {
				addLog(x)
			}
			walk.MsgBox(mw, "Результат диагностики", strings.Join(lines, "\r\n"), walk.MsgBoxIconInformation)
		})
	}()
}

func run() error {
	return (MainWindow{
		AssignTo: &mw, Title: "BELHUB 2.2 — госпорталы Республики Беларусь", MinSize: Size{Width: 900, Height: 650}, Size: Size{Width: 1120, Height: 760},
		Layout: VBox{MarginsZero: false, Spacing: 10},
		MenuItems: []MenuItem{
			Menu{Text: "Файл", Items: []MenuItem{Action{Text: "Выход", OnTriggered: func() { mw.Close() }}}},
			Menu{Text: "Инструменты", Items: []MenuItem{Action{Text: "Диагностика", OnTriggered: diagnostics}, Action{Text: "Официальные загрузки АВЕСТ", OnTriggered: func() { openDefault("https://www.avest.by/crypto/download/") }}, Action{Text: "Обновление сертификатов и СОС", OnTriggered: func() { openDefault("https://helpdesk.nces.by/client-docs/avest/instrukciya-po-ustanovke-avest") }}}},
			Menu{Text: "Справка", Items: []MenuItem{Action{Text: "О программе", OnTriggered: func() {
				walk.MsgBox(mw, "О программе", "BELHUB 2.2\r\nБезопасный каталог госпорталов Республики Беларусь.\r\n\r\nПрограмма не хранит учетные данные и ключи ЭЦП.", walk.MsgBoxIconInformation)
			}}}},
		},
		StatusBarItems: []StatusBarItem{{AssignTo: &status, Text: "Готово"}},
		Children: []Widget{
			Composite{Layout: HBox{Spacing: 8}, Children: []Widget{
				Label{Text: "Поиск:"}, LineEdit{AssignTo: &search, OnTextChanged: refresh},
				Label{Text: "Категория:"}, ComboBox{AssignTo: &category, Model: categories(), CurrentIndex: 0, OnCurrentIndexChanged: refresh},
				PushButton{Text: "Сбросить", OnClicked: func() { search.SetText(""); category.SetCurrentIndex(0); refresh() }},
			}},
			HSplitter{Children: []Widget{
				GroupBox{Title: "Порталы", Layout: VBox{}, Children: []Widget{ListBox{AssignTo: &list, OnCurrentIndexChanged: showDetails, OnItemActivated: func() {
					if p, ok := selected(); ok {
						openDefault(p.URL)
					}
				}}}},
				Composite{Layout: VBox{Spacing: 9}, Children: []Widget{
					GroupBox{Title: "Информация и вход", Layout: VBox{}, Children: []Widget{TextEdit{AssignTo: &details, ReadOnly: true, VScroll: true}}},
					PushButton{Text: "Открыть портал", MinSize: Size{Height: 38}, OnClicked: func() {
						if p, ok := selected(); ok {
							openDefault(p.URL)
						}
					}},
					PushButton{Text: "Открыть в Microsoft Edge", OnClicked: func() {
						if p, ok := selected(); ok {
							openEdge(p.URL)
						}
					}},
					PushButton{Text: "Добавить портал в «Надежные сайты»", OnClicked: addTrusted},
					HSeparator{},
					PushButton{Text: "Диагностика рабочего места", OnClicked: diagnostics},
					PushButton{Text: "Скачать / установить АВЕСТ с официального сайта", OnClicked: func() { openDefault("https://www.avest.by/crypto/download/") }},
					PushButton{Text: "Обновить сертификаты и списки отзыва (СОС)", OnClicked: func() { openDefault("https://helpdesk.nces.by/client-docs/avest/instrukciya-po-ustanovke-avest") }},
				}},
			}},
			GroupBox{Title: "Журнал", Layout: VBox{}, Children: []Widget{TextEdit{AssignTo: &logBox, ReadOnly: true, VScroll: true, MinSize: Size{Height: 105}}}},
		},
	}.Create())
}

func main() {
	if err := run(); err != nil {
		walk.MsgBox(nil, "BELHUB — ошибка запуска", err.Error(), walk.MsgBoxIconError)
		return
	}
	addLog("BELHUB 2.2 запущен")
	refresh()
	mw.Run()
}
