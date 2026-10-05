# BELBUH

Репозиторий проекта BELHUB/BELBUH.

## Основная сборка BELHUB 3.2

Актуальные исходники для автоматической сборки находятся в папке [`BELBUH-repo/`](BELBUH-repo/):

- HTML-приложение: `BELBUH-repo/web/`;
- Windows launcher: `BELBUH-repo/cmd/launcher/`;
- launcher мини-виджета: `BELBUH-repo/cmd/widget/`;
- NSIS-установщик: `BELBUH-repo/installer/installer.nsi`;
- скрипт сборки: `BELBUH-repo/build.ps1`.

## GitHub Actions

Автоматическая сборка настроена в [`.github/workflows/build.yml`](.github/workflows/build.yml).

Workflow **Build BELHUB** запускается после push в GitHub, для pull request в `main`, вручную через вкладку **Actions**, а также при публикации тегов вида `v*`.

Подробная документация и инструкции для следующего агента:

- [`docs/CI_AUTOMATION.md`](docs/CI_AUTOMATION.md) — сборка и GitHub Actions;
- [`docs/PORTAL_LINKS_AUDIT.md`](docs/PORTAL_LINKS_AUDIT.md) — сверка ссылок с BELPORTAL и логика AvTunProxy.
