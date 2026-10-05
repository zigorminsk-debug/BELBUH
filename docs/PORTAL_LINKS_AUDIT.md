# Сверка порталов BELHUB с BELPORTAL

Дата сверки: 2026-10-05.

Рабочая ветка Arena: `arena/01a10ba9-belbuh`.

## Источники сверки

1. Страница BELPORTAL: `https://belportal.by/`.
2. Страница загрузки BELPORTAL: `https://belportal.by/belPortalDonwload`.
3. Локально сохранённый ClickOnce-пакет BELPORTAL из репозитория:
   - `BelPortal.application`, версия `2.9.2.10`;
   - `BelPortal.exe.deploy`.
4. Строки URL из `BelPortal.exe.deploy`, полученные командой:

   ```bash
   strings -el BelPortal.exe.deploy | grep -Eoi 'https?://[^" <>]+' | sort -u
   ```

## Найденные URL BELPORTAL, учтённые в BELHUB

| Назначение | URL |
| --- | --- |
| МНС | `http://portal.nalog.gov.by/` |
| ЭСЧФ / VAT | `http://vat.gov.by/` |
| ФСЗН, новый корпоративный портал | `https://portal2.ssf.gov.by/mainPage/?step=payer` |
| ФСЗН, старый портал | `http://portal.ssf.gov.by/mainPage/?step=2` |
| Электронный респондент Белстата | `http://e-respondent.belstat.gov.by/belstat/` |
| Регистрация валютных договоров | `https://rvd.nbrb.by/nbrbResidentUi/#/` |
| ЕГР | `https://egr.gov.by/egrn/egrportal.html?type=id-card` |
| Белгосстрах — отчётность | `https://reporting.bgs.by/` |
| ЭДиН | `https://app.edn.by/web-new/login` |
| ЭДиН CloudSign | `https://app.edn.by/cloudsign/download` |
| Bidmart EDI | `https://edi.bidmart.by/user/login` |
| ilex.Накладные | `https://edi.ilex.by/login` |
| СТТ | `https://ctt.by/` |
| Подпис.бел | `https://app.podpis.by/` |
| СККО | `http://skko.by/` |
| СККО — кабинет сервисного центра | `https://lk-sa.skko.by:2222/` |
| DataMark | `https://i.datamark.by/auth` |
| iKassa | `https://lk.ikassa.by/auth` |
| Работа в Беларуси | `https://gsz.gov.by/` |
| Минскэнерго | `https://lkyl.minskenergo.by/authorization/login` |
| Брестэнерго | `https://lk-enterprises.brestenergo.by/authorization/login` |
| Гомельэнерго | `https://lkul.gomelenergo.by/authorization/login` |
| Витебскэнерго | `https://lkyl.vitebskenergo.by/authorization/login` |
| Гродноэнерго | `https://legalcab.energo.grodno.by/authorization/login` |
| Минскводоканал | `https://lk.minskvodokanal.by/` |

Служебные URL BELPORTAL, новости, API, изображения, AnyDesk и файлы обновлений самого BELPORTAL в каталог порталов не переносились.

## Что изменено в BELHUB

1. В `BELBUH-repo/web/index.html` обновлён массив `portals`.
2. Добавлен портал ЭСЧФ / VAT по строке BELPORTAL `http://Vat.gov.by/`; в BELHUB URL нормализован до `http://vat.gov.by/`.
3. Исправлены ссылки на порталы, которые в BELPORTAL открываются не с корня домена:
   - ФСЗН: `https://portal2.ssf.gov.by/mainPage/?step=payer`;
   - старый ФСЗН: `http://portal.ssf.gov.by/mainPage/?step=2`;
   - Белстат-респондент: `http://e-respondent.belstat.gov.by/belstat/`;
   - ЕГР: `https://egr.gov.by/egrn/egrportal.html?type=id-card`;
   - Правовой портал: раздел кодексов.
4. Добавлены дополнительные порталы из BELPORTAL:
   - регистрация валютных договоров;
   - Белгосстрах — отчётность;
   - Подпис.бел;
   - СККО;
   - DataMark;
   - iKassa;
   - Работа в Беларуси;
   - кабинеты энергоснабжающих организаций;
   - Минскводоканал.
5. В `BELBUH-repo/web/widget.html` добавлены быстрые плитки:
   - личный кабинет МНС;
   - ЭСЧФ / VAT;
   - регистрация валютных договоров;
   - обновлённый вход ФСЗН.
6. В меню **ПО** добавлен пункт **AvTunProxy** с оригинальной ссылкой:
   - `https://avtunproxy.by/windows/`.

## Проверка AvTunProxy

Для порталов:

- ФСЗН — корпоративный портал;
- ФСЗН — старый портал;
- Регистрация валютных договоров;

BELHUB теперь не открывает адрес напрямую из HTML. Вместо этого используется локальный протокол:

```text
belhub://avtun?target=<url-encoded portal url>
```

Обработчик находится в `BELBUH-repo/cmd/launcher/main.go`.

Для безопасности обработчик принимает только целевые хосты `portal2.ssf.gov.by`, `portal.ssf.gov.by` и `rvd.nbrb.by`; произвольные URL через `belhub://avtun` не открываются.

Алгоритм обработчика:

1. Проверить, отвечает ли локальный PAC AvTunProxy:

   ```text
   http://127.0.0.1:10224/proxy.pac
   ```

2. Если PAC отвечает — открыть целевой портал.
3. Если PAC не отвечает — попытаться найти установленный `AvTunProxy.exe` в типовых местах:
   - `%ProgramFiles%\AvTunProxy\AvTunProxy.exe`;
   - `%ProgramFiles(x86)%\AvTunProxy\AvTunProxy.exe`;
   - `%ProgramFiles%\Avest\AvTunProxy\AvTunProxy.exe`;
   - `%ProgramFiles(x86)%\Avest\AvTunProxy\AvTunProxy.exe`;
   - `%ProgramFiles%\Avest\AvTun\AvTunProxy.exe`;
   - `%ProgramFiles(x86)%\Avest\AvTun\AvTunProxy.exe`;
   - `%LOCALAPPDATA%\Programs\AvTunProxy\AvTunProxy.exe`;
   - `%LOCALAPPDATA%\AvTunProxy\AvTunProxy.exe`;
   - `PATH` через `exec.LookPath("AvTunProxy.exe")`.
4. Если файл найден — запустить его, коротко подождать PAC и открыть портал.
5. Если AvTunProxy не найден — показать пользователю диалог и предложить скачать установщик с:

   ```text
   https://avtunproxy.by/windows/
   ```

## Регистрация протокола `belhub://`

Чтобы проверка AvTunProxy работала не только после запуска основного `BELHUB-3.2.exe`, протокол регистрируется в трёх местах:

1. При запуске основного launcher:
   - `BELBUH-repo/cmd/launcher/main.go`.
2. При запуске виджета, если рядом найден `BELHUB-3.2.exe`:
   - `BELBUH-repo/cmd/widget/main.go`.
3. При установке через NSIS:
   - `BELBUH-repo/installer/installer.nsi`.

При удалении через NSIS ключ `HKCU\Software\Classes\belhub` удаляется.

## Проверка после изменений

В этой Linux-среде Go не установлен, поэтому локальная кросс-компиляция невозможна. Минимальная проверка, выполненная здесь:

```bash
node /tmp/belhub-script.js
```

Скрипт из `BELBUH-repo/web/index.html` успешно скомпилирован через `new Function(...)`, то есть синтаксических ошибок JavaScript не обнаружено.

Полная проверка должна выполняться GitHub Actions на Windows:

```bash
git push origin arena/01a10ba9-belbuh
```

Далее проверить workflow **Build BELHUB** в GitHub Actions.

## Что проверить следующему агенту

1. Дождаться успешного GitHub Actions после push.
2. Скачать artifact и проверить в Windows:
   - открытие ЭСЧФ / VAT (`http://vat.gov.by/`), так как ссылка присутствует в BELPORTAL;
   - открытие ФСЗН с установленным AvTunProxy;
   - диалог предложения скачать AvTunProxy, если компонент отсутствует;
   - открытие регистрации валютных договоров с установленным AvTunProxy;
   - виджет: плитки ФСЗН и валютных договоров должны вызывать `belhub://avtun`.
3. Если официальный список BELPORTAL изменится, повторить извлечение URL из нового `BelPortal.exe.deploy` и обновить таблицу выше.
