# iCalendar (Parser & Encoder)

> [EN](README.md)
> RU

![Go Version](docs/assets/go.svg)
[![RFC 5545](docs/assets/rfc.svg)](docs/ical_analysis.md#анализ-rfc-5545-icalendar-и-компонента-vevent)
[![RFC 7986](docs/assets/rfc-7986.svg)](docs/rfc_extensions_analysis.md#1-rfc-7986-новые-свойства-для-данных-icalendar)
[![RFC 7953](docs/assets/rfc-7953.svg)](docs/rfc_extensions_analysis.md#2-rfc-7953-доступность-календаря-vavailability)
[![RFC 9073](docs/assets/rfc-9073.svg)](docs/rfc_extensions_analysis.md#3-rfc-9073-расширения-для-публикации-событий)
[![RFC 9253](docs/assets/rfc-9253.svg)](docs/rfc_extensions_analysis.md#4-rfc-9253-поддержка-взаимосвязей-в-icalendar)
[![RFC 9074](docs/assets/rfc-9074.svg)](docs/rfc_extensions_analysis.md#5-rfc-9074-расширения-valarm)
[![RFC 7529](docs/assets/rfc-7529.svg)](docs/rfc_extensions_analysis.md#6-rfc-7529-негригорианские-rrule)
[![RFC 6868](docs/assets/rfc-6868.svg)](docs/rfc_extensions_analysis.md)

Высокопроизводительная библиотека для работы с форматом **iCalendar (RFC 5545)** на языке Go. Позволяет парсить, конструировать и сериализовать календари, обеспечивая скорость работы, минимальное потребление памяти и строгую типизацию.

Отлично подходит для обработки и генерации объемных календарей (например, `.ics` файлов для Google Calendar, Apple Calendar, Microsoft Outlook и других провайдеров).

---

## Особенности (Features)

- ⚡️ **Высокая производительность:** Потоковый парсинг через `io.Reader` и запись через `io.Writer` с минимальными аллокациями.
- 🛠 **Builder Pattern:** Удобное создание календарей и событий через функциональные опции (Functional Options).
- 📤 **Encoder:** Глубокая сериализация моделей Go обратно в валидный формат `.ics` (RFC 5545).
- 📆 **Поддержка времени и зон:** Интеграция со стандартным пакетом `time`, работа с `TZID` и компонентом `VTIMEZONE`.
- 🔁 **Сложные повторения (Recurrence):** Парсинг и генерация `RRULE`, `RDATE` и `EXDATE` (Daily, Weekly, сложные комбинации BYDAY и т.д.).
- 🛡 **Устойчивость к ошибкам:** Корректная обработка vendor-specific свойств (X-Props), файлов с `LF` и смешанными `CRLF`/`LF`.
- 📦 **Понятная объектная модель:** Чистые структуры данных, независимые от исходного текста. Перечисления для константных параметров.
- 🧩 **Поддержка расширений RFC:** `RFC 6868`, `RFC 7529`, `RFC 7953`, `RFC 7986`, `RFC 9073`, `RFC 9074`, `RFC 9253`.

---

## Поддержка стандарта RFC 5545

Парсер и энкодер поддерживают наиболее востребованные компоненты стандарта. Благодаря расширяемой архитектуре, даже нестандартные свойства сохраняются и могут быть обработаны.

### Компоненты (Components)

| Компонент   | Статус | Описание                                                                  |
| ----------- | :----: | ------------------------------------------------------------------------- |
| `VCALENDAR` |   🟢    | Корневой объект, поддержка `VERSION`, `PRODID`, `CALSCALE`, `METHOD`      |
| `VEVENT`    |   🟢    | Календарное событие: полная поддержка (даты, статусы, участники, правила) |
| `VTIMEZONE` |   🟢    | Часовые пояса: переходное время `STANDARD` и `DAYLIGHT`                   |
| `VALARM`    |   🟢    | Оповещения и будильники: триггеры (`TRIGGER`), действия, длительности     |
| `VFREEBUSY` |   🟢    | Информация о занятости: `FREE`, `BUSY`, списки периодов (Periods)         |
| `VTODO`     |   🟢    | Задачи: дедлайны (`DUE`), прогресс, приоритеты, вложенные оповещения      |
| `VJOURNAL`  |   🟢    | Записи в журнале: полная поддержка текстовых описаний и статусов          |

### Свойства событий (VEVENT Properties)

| Свойство          | Статус | Комментарий                                                              |
| ----------------- | :----: | ------------------------------------------------------------------------ |
| **Даты и время**  |   🟢    | `DTSTART`, `DTEND`, `DURATION`, `DTSTAMP`, `CREATED`, `LAST-MODIFIED`    |
| **Идентификация** |   🟢    | `UID`, `SEQUENCE`, `RELATED-TO`, `RECURRENCE-ID`                         |
| **Описание**      |   🟢    | `SUMMARY`, `DESCRIPTION`, `LOCATION`, `URL`                              |
| **Участники**     |   🟢    | `ORGANIZER`, `ATTENDEE` (с параметрами `ROLE`, `PARTSTAT`, `RSVP` и др.) |
| **Повторения**    |   🟢    | `RRULE`, `RDATE`, `EXDATE`                                               |
| **Статусы**       |   🟢    | `STATUS`, `TRANSP` (`OPAQUE`/`TRANSPARENT`), `CLASS`                     |
| **Прочее**        |   🟢    | Vendor-specific расширения (через `XProps`)                              |

---

## Поддержка расширений iCalendar

Поддерживаются ключевые RFC-расширения для iCalendar. Подробности и примеры — в [`docs/rfc_extensions_analysis.md`](docs/rfc_extensions_analysis.md).

| RFC          | Статус | Что покрыто                                                                                                                                                               |
| ------------ | :----: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **RFC 6868** |   🟢    | `^`-экранирование в значениях параметров: `^n` → `\n`, `^^` → `^`, `^'` → `"` — поддержка в parser, encoder и builder                                                     |
| **RFC 7529** |   🟢    | `RSCALE`, `SKIP` в правилах `RRULE`                                                                                                                                       |
| **RFC 7953** |   🟢    | `VAVAILABILITY` и `AVAILABLE`, поддержка рекуррентных окон доступности                                                                                                    |
| **RFC 7986** |   🟢    | Свойства `VCALENDAR`: `NAME`, `DESCRIPTION`, `UID`, `LAST-MODIFIED`, `URL`, `CATEGORIES`, `REFRESH-INTERVAL`, `COLOR`, `SOURCE`; IANA свойства RFC 7986 через `IanaProps` |
| **RFC 9073** |   🟢    | `STRUCTURED-DATA`, `STYLED-DESCRIPTION`, компоненты `PARTICIPANT`, `LOCATION`, `RESOURCE`                                                                                 |
| **RFC 9074** |   🟢    | `ACKNOWLEDGED`, `PROXIMITY`, `SNOOZE`, `VLOCATION` для `VALARM`                                                                                                           |
| **RFC 9253** |   🟢    | `LINK`, `CONCEPT`, `REFID`, расширения `RELATED-TO`                                                                                                                       |

---

## Установка

Используйте стандартный инструментарий `go get`:

```bash
go get github.com/mscloudx/ical
```

---

## Быстрый старт (Quick Start)

Пакет `parser` предоставляет удобные точки входа: `Parse(io.Reader)`, `ParseBytes([]byte)` и `ParseWithClose(io.ReadCloser)`.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mscloudx/ical/parser"
)

func main() {
	f, err := os.Open("calendar.ics")
	if err != nil {
		log.Fatalf("Ошибка при открытии файла: %v", err)
	}
	defer f.Close()

	// Парсинг потока
	calendar, err := parser.Parse(f)
	if err != nil {
		log.Fatalf("Ошибка разбора iCalendar: %v", err)
	}

	fmt.Printf("Календарь (Версия: %s)\n", calendar.Version)
	
	// Вывод событий
	for _, event := range calendar.Events {
		fmt.Printf("- Событие: %s\n", event.Summary)
		fmt.Printf("  Начало: %s\n", event.DTStart.Format("2006-01-02 15:04:05"))
	}
}
```

### Создание и экспорт (Builder & Encoder)

Библиотека предоставляет удобный способ создания новых календарей «с нуля» и их сохранения в файл.

```go
package main

import (
	"os"
	"time"

	"github.com/mscloudx/ical/builder"
	"github.com/mscloudx/ical/encoder"
)

func main() {
	// Создаем событие с помощью билдера
	now := time.Now()
	event := builder.NewEvent("unique-id@example.com", now, now.Add(2*time.Hour),
		builder.WithEventSummary("Встреча с командой"),
		builder.WithEventDescription("Обсуждение планов на неделю"),
		builder.WithEventLocation("Room 101"),
	)

	// Собираем календарь
	cal := builder.NewCalendar("-//mscloudx//iCalendar//EN",
		builder.WithEvents(event),
	)

	// Сериализуем в файл
	f, _ := os.Create("my_calendar.ics")
	defer f.Close()

	encoder.Encode(f, cal)
}
```

---

## Архитектура и Документация

Проект следует принципам **Clean Architecture** (в рамках Go-way). Библиотека разделена на логические пакеты, чтобы избежать циклических зависимостей и упростить тестирование:

- `model/` — Доменные структуры данных (календари, события, типы, перечисления).
- `parser/` — Логика разбора iCalendar-потока, лексер и парсинг сложных свойств.
- `builder/` — Fluent-интерфейс для конструирования моделей через Functional Options.
- `encoder/` — Сериализация моделей в текстовый формат RFC 5545 с поддержкой фолдинга строк.

Обязательно загляните в папку [`docs/`](docs/) для углубления в детали стандарта:
- [📖 Глоссарий терминов (glossary.md)](docs/glossary.md)
- [⚙️ Анализ RFC 5545 (ical_analysis.md)](docs/ical_analysis.md)
- [📈 Расширения iCalendar (rfc_extensions_analysis.md)](docs/rfc_extensions_analysis.md)
- [🛠 Особенности вендоров (vendor_quirks.md)](docs/vendor_quirks.md)

---

## Тестирование и Качество кода

Библиотека покрыта тестами с использованием парадигмы AAA (Arrange, Act, Assert) и `testify/suite`.
Для проверки работы с реальными календарями в репозитории добавлена обширная папка `examples/ical/`, покрывающая как стандартные случаи, так и "кривые" файлы от конкретных провайдеров.

Для запуска тестов, линтинга и форматирования:
```bash
make test
make lint
make fmt
```

Также доступны прямые команды:
```bash
go test ./...
go test -bench=. ./parser/
```

Код соответствует строгим стандартам: используется `golangci-lint`, форматирование через `gofumpt` и импорт-органайзер `gci`. Все комментарии к публичному API и коду выполнены на **русском языке**.

---

## Лицензия (License)

Этот проект распространяется под лицензией **MIT License**. Вы можете свободно использовать, изменять и распространять код для любых (в том числе коммерческих) целей при условии указания авторства. Подробнее см. в файле [LICENSE](LICENSE).

---

*Примечание: для аналитики, составления документации и частичного написания кода проекта использовался искусственный интеллект.*
