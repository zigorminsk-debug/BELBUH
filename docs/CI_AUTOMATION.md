# Автоматическая сборка BELHUB через GitHub Actions

Дата настройки: 2026-10-05.

Рабочая ветка Arena: `arena/01a10ba9-belbuh`.

Этот документ описывает, что было сделано для CI-сборки, как запустить её вручную и как следующему агенту продолжить работу без потери контекста.

## Что настроено

Добавлен workflow GitHub Actions:

- файл: `.github/workflows/build.yml`;
- имя в интерфейсе GitHub: **Build BELHUB**;
- основная папка исходников: `BELBUH-repo/`;
- локальный скрипт сборки, который вызывает workflow: `BELBUH-repo/build.ps1`.

Workflow собирает Windows-версию BELHUB 3.2, формирует установщик NSIS и публикует результат как artifact GitHub Actions.

## Триггеры запуска

Workflow запускается автоматически:

1. При `push` в любую ветку, если изменились:
   - `BELBUH-repo/**`;
   - `.github/workflows/build.yml`;
   - `docs/**`;
   - корневой `README.md`.
2. При `pull_request` в `main` с теми же фильтрами путей.
3. Вручную из GitHub: **Actions → Build BELHUB → Run workflow**.
4. При публикации Git-тега вида `v*`, например `v3.2.1`. Для такого тега дополнительно создаётся GitHub Release.

## Что делает workflow

Пайплайн выполняется на `windows-latest` и проходит такие шаги:

1. `actions/checkout@v4` — получает код репозитория.
2. `actions/setup-go@v5` — устанавливает Go `1.25.x` и включает cache для модулей из:
   - `BELBUH-repo/cmd/launcher/go.mod`;
   - `BELBUH-repo/cmd/widget/go.mod`.
3. Устанавливает компилятор Windows-ресурсов:
   - `go install github.com/akavel/rsrc@latest`.
4. Устанавливает NSIS через Chocolatey:
   - `choco install nsis --yes --no-progress`.
5. Добавляет каталог NSIS в `PATH`, чтобы был доступен `makensis.exe`.
6. Запускает сборку:
   - `.\BELBUH-repo\build.ps1`.
7. Создаёт файл контрольных сумм:
   - `BELBUH-repo/dist/SHA256SUMS.txt`.
8. Загружает artifact с результатами сборки на 30 дней.
9. Если workflow запущен тегом `v*`, создаёт GitHub Release и прикладывает установщик с SHA-256 суммами.

## Результаты сборки

После успешного запуска файлы доступны в GitHub:

**Actions → Build BELHUB → нужный запуск → Artifacts**

Состав artifact:

- `BELHUB-3.2-Setup.exe` — установщик;
- `BELHUB-3.2.exe` — основной launcher;
- `BELHUB-Widget.exe` — launcher виджета;
- `BELHUB-3.2.html` — основное HTML-приложение;
- `widget.html` — HTML виджета;
- `SHA256SUMS.txt` — SHA-256 контрольные суммы.

## Релиз по тегу

Чтобы выпустить релиз через GitHub Actions:

```bash
git tag v3.2.1
git push origin v3.2.1
```

Правило имени тега: он должен начинаться с `v`, например `v3.2.1` или `v3.3.0`.

При запуске по тегу workflow создаст GitHub Release через `softprops/action-gh-release@v2` и приложит:

- `BELBUH-repo/dist/BELHUB-3.2-Setup.exe`;
- `BELBUH-repo/dist/SHA256SUMS.txt`.

## Исходники, которые нужны для сборки

Для работоспособности сборки в `BELBUH-repo/` должны присутствовать:

```text
BELBUH-repo/
├── assets/
│   ├── BELHUB.ico
│   ├── app-icon.png
│   └── csl-logo.png
├── cmd/
│   ├── launcher/
│   │   ├── app.manifest
│   │   ├── go.mod
│   │   └── main.go
│   └── widget/
│       ├── go.mod
│       └── main.go
├── installer/
│   └── installer.nsi
├── web/
│   ├── index.html
│   └── widget.html
├── build.ps1
└── README.md
```

Папки `BELBUH-repo/build/` и `BELBUH-repo/dist/` являются временными результатами сборки и не должны попадать в Git.

## Локальная сборка на Windows

Требования:

- Windows;
- PowerShell;
- Go;
- NSIS 3 (`makensis.exe` должен быть в `PATH`);
- `rsrc` — если отсутствует, `build.ps1` установит его через `go install github.com/akavel/rsrc@latest`.

Команда из корня репозитория:

```powershell
powershell -ExecutionPolicy Bypass -File .\BELBUH-repo\build.ps1
```

Либо из папки исходников:

```powershell
cd BELBUH-repo
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

Результаты будут в `BELBUH-repo/dist/`.

## Быстрая проверка без Windows/NSIS

В Linux-среде нельзя полноценно проверить NSIS-установщик, но можно проверить, что Go-исходники компилируются под Windows:

```bash
cd BELBUH-repo/cmd/launcher
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H windowsgui' -o /tmp/BELHUB-3.2.exe .

cd ../widget
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H windowsgui' -o /tmp/BELHUB-Widget.exe .
```

Это не заменяет полный запуск workflow, потому что не проверяет генерацию `.syso` через `rsrc` и сборку установщика через NSIS.

## Что было изменено при настройке

1. Добавлен `.github/workflows/build.yml`.
2. Восстановлены исходники launcher-компонентов в `BELBUH-repo/cmd/`:
   - `cmd/launcher`;
   - `cmd/widget`.
3. Добавлена документация `docs/CI_AUTOMATION.md`.
4. Обновлён корневой `README.md` как навигационная точка для следующих агентов.
5. Обновлена секция автоматической сборки в `BELBUH-repo/README.md`.

## Порядок действий для следующего агента

1. Всегда начинать с проверки состояния:

   ```bash
   git status --short --branch
   ```

2. Не переключаться с ветки Arena `arena/01a10ba9-belbuh`, если работа идёт в этой сессии.
3. После изменений в приложении проверять минимум:

   ```bash
   cd BELBUH-repo/cmd/launcher
   GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H windowsgui' -o /tmp/BELHUB-3.2.exe .

   cd ../widget
   GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -H windowsgui' -o /tmp/BELHUB-Widget.exe .
   ```

4. Для полной проверки дождаться GitHub Actions после `push` или запустить workflow вручную.
5. Если workflow упал:
   - открыть лог упавшего шага в **Actions → Build BELHUB**;
   - сначала проверить наличие `rsrc`, `makensis.exe`, файлов в `BELBUH-repo/dist/`;
   - затем проверить, не изменились ли имена выходных файлов в `build.ps1` или `installer.nsi`.
6. При изменении версии BELHUB не забыть синхронизировать версию в:
   - `BELBUH-repo/build.ps1`;
   - `BELBUH-repo/installer/installer.nsi`;
   - `BELBUH-repo/cmd/launcher/app.manifest`;
   - `BELBUH-repo/README.md`;
   - `.github/workflows/build.yml`, если меняются имена artifact-файлов.

## Замечания

- Для сборки не нужны дополнительные GitHub Secrets.
- Разрешение `contents: write` нужно только для публикации GitHub Release по тегу.
- Artifact хранится 30 дней. Для долговременного хранения используйте GitHub Releases.
- Workflow не коммитит собранные `.exe` обратно в репозиторий — это сделано намеренно, чтобы не засорять историю бинарными файлами.
