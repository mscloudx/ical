# CloudCoder iCalendar (Parser & Encoder)

![Go Version](docs/assets/go.svg)
![RFC 5545](docs/assets/rfc.svg)

Высокопроизводительная библиотека для работы с форматом **iCalendar (RFC 5545)** на языке Go. Позволяет парсить, конструировать и сериализовать календари, обеспечивая скорость работы, минимальное потребление памяти и строгую типизацию.

Отлично подходит для обработки и генерации объемных календарей (например, `.ics` файлов для Google Calendar, Apple Calendar, Microsoft Outlook и других провайдеров).

---

## Особенности (Features)

- ⚡️ **Высокая производительность:** Потоковый парсинг через `io.Reader` и запись через `io.Writer` с минимальными аллокациями.
- 🛠 **Builder Pattern:** Удобное создание календарей и событий через функциональные опции (Functional Options).
- 📤 **Encoder:** Глубокая сериализация моделей Go обратно в валидный формат `.ics` (RFC 5545).
- 📆 **Полная поддержка времени и зон:** Интеграция со стандартным пакетом `time`, автоматический резолвинг `TZID` (часовых поясов).
- 🔁 **Сложные повторения (Recurrence):** Парсинг и генерация `RRULE`, `RDATE` и `EXDATE` (Daily, Weekly, сложные комбинации BYDAY и т.д.).
- 🛡 **Устойчивость к ошибкам:** Корректная обработка vendor-specific свойств (X-Props), файлов с `LF` и смешанными `CRLF`/`LF`.
- 📦 **Понятная объектная модель:** Чистые структуры данных, независимые от исходного текста. Перечисления для константных параметров.

---

## Поддержка стандарта RFC 5545

Парсер и энкодер поддерживают наиболее востребованные компоненты стандарта. Благодаря расширяемой архитектуре, даже нестандартные свойства сохраняются и могут быть обработаны.

### Компоненты (Components)

| Компонент      | Статус | Описание |
|----------------|:---:|----------|
| `VCALENDAR`    | 🟢 | Корневой объект, поддержка `VERSION`, `PRODID`, `CALSCALE`, `METHOD` |
| `VEVENT`       | 🟢 | Календарное событие: полная поддержка (даты, статусы, участники, правила) |
| `VTIMEZONE`    | 🟢 | Часовые пояса: переходное время `STANDARD` и `DAYLIGHT` |
| `VALARM`       | 🟢 | Оповещения и будильники: триггеры (`TRIGGER`), действия, длительности |
| `VFREEBUSY`    | 🟢 | Информация о занятости: `FREE`, `BUSY`, списки периодов (Periods) |
| `VTODO`        | 🟢 | Задачи: дедлайны (`DUE`), прогресс, приоритеты, вложенные оповещения |
| `VJOURNAL`     | 🟢 | Записи в журнале: полная поддержка текстовых описаний и статусов |

### Свойства событий (VEVENT Properties)

| Свойство | Статус | Комментарий |
|---|:---:|---|
| **Даты и время** | 🟢 | `DTSTART`, `DTEND`, `DURATION`, `DTSTAMP`, `CREATED`, `LAST-MODIFIED` |
| **Идентификация**| 🟢 | `UID`, `SEQUENCE`, `RELATED-TO`, `RECURRENCE-ID` |
| **Описание** | 🟢 | `SUMMARY`, `DESCRIPTION`, `LOCATION`, `URL` |
| **Участники** | 🟢 | `ORGANIZER`, `ATTENDEE` (с параметрами `ROLE`, `PARTSTAT`, `RSVP` и др.) |
| **Повторения** | 🟢 | `RRULE`, `RDATE`, `EXDATE` |
| **Статусы** | 🟢 | `STATUS`, `TRANSP` (`OPAQUE`/`TRANSPARENT`), `CLASS` |
| **Прочее** | 🟢 | Vendor-specific расширения (через `XProps`) |

---

## Установка

Используйте стандартный инструментарий `go get`:

```bash
go get gitverse.ru/cloudcoder/ical
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

	"gitverse.ru/cloudcoder/ical/parser"
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

	"gitverse.ru/cloudcoder/ical/builder"
	"gitverse.ru/cloudcoder/ical/encoder"
)

func main() {
	// Создаем событие с помощью билдера
	event := builder.NewEvent("unique-id@example.com", time.Now(),
		builder.WithEventSummary("Встреча с командой"),
		builder.WithEventDescription("Обсуждение планов на неделю"),
		builder.WithEventLocation("Room 101"),
	)

	// Собираем календарь
	cal := builder.NewCalendar("-//CloudCoder//iCalendar//EN",
		builder.WithEvents(*event),
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
- [🛠 Особенности вендоров (vendor_quirks.md)](docs/vendor_quirks.md)

---

## Тестирование и Качество кода

Библиотека покрыта тестами с использованием парадигмы AAA (Arrange, Act, Assert) и `testify/suite`.
Для проверки работы с реальными календарями в репозитории добавлена обширная папка `examples/ical/`, покрывающая как стандартные случаи, так и "кривые" файлы от конкретных провайдеров.

Для запуска тестов и бенчмарков:
```bash
go test ./...
go test -bench=. ./parser/
```

Код соответствует строгим стандартам, настроен `golangci-lint` (v2). Все комментарии к публичному API и коду выполнены на **русском языке**.

---

## Лицензия (License)

Этот проект распространяется под лицензией **MIT License**. Вы можете свободно использовать, изменять и распространять код для любых (в том числе коммерческих) целей при условии указания авторства. Подробнее см. в файле [LICENSE](LICENSE).

---

*Примечание: для аналитики, составления документации и частичного написания кода проекта использовался искусственный интеллект.*
