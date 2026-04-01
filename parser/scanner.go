package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/mscloudx/ical/model"
)

// defaultBufSize — размер буфера для чтения физических строк.
const defaultBufSize = 8192

// lineScanner читает физические строки из io.Reader и выполняет
// unfolding (RFC 5545 §3.1): склеивает строки-продолжения,
// начинающиеся с пробела или табуляции.
type lineScanner struct {
	r    *bufio.Reader
	buf  bytes.Buffer
	err  error
	done bool
}

// newLineScanner создаёт сканер с буферизацией.
func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{
		r: bufio.NewReaderSize(r, defaultBufSize),
	}
}

// Scan считывает следующую логическую (развёрнутую) строку.
// Возвращает false, когда данные закончились или произошла ошибка.
func (s *lineScanner) Scan() bool {
	if s.done {
		return false
	}

	s.buf.Reset()

	// Читаем первую физическую строку.
	line, err := s.readLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			s.done = true
			if len(line) > 0 {
				s.buf.Write(line)
				return true
			}
			return false
		}
		s.err = err
		s.done = true
		return false
	}
	s.buf.Write(line)

	// Разворачиваем строки-продолжения (unfolding).
	for {
		b, err := s.r.Peek(1)
		if err != nil || (b[0] != ' ' && b[0] != '\t') {
			break
		}
		cont, err := s.readLine()
		// Пропускаем ведущий пробел/табуляцию.
		// Данные записываем до проверки ошибки: ReadBytes возвращает
		// прочитанные байты вместе с io.EOF при отсутствии финального \n.
		if len(cont) > 0 {
			s.buf.Write(cont[1:])
		}
		if err != nil {
			break
		}
	}

	return true
}

// Bytes возвращает текущую логическую строку.
func (s *lineScanner) Bytes() []byte {
	return s.buf.Bytes()
}

// Err возвращает ошибку сканирования (кроме io.EOF).
func (s *lineScanner) Err() error {
	return s.err
}

// readLine читает одну физическую строку, обрезая CRLF и LF.
func (s *lineScanner) readLine() ([]byte, error) {
	line, err := s.r.ReadBytes('\n')
	// Обрезаем завершающие \n и \r.
	line = bytes.TrimRight(line, "\r\n")
	return line, err
}

// --------------------------------------------------------------------------
// Парсинг content-line: имя, параметры, значение
// --------------------------------------------------------------------------

// parseContentLine разбирает развёрнутую строку на Property.
//
// Формат: NAME[;PARAM1=VAL1;PARAM2=VAL2]:VALUE
func parseContentLine(line []byte) model.Property {
	var p model.Property

	// Ищем имя свойства: до первого ';' или ':'.
	i := 0
	for i < len(line) && line[i] != ';' && line[i] != ':' {
		i++
	}
	p.Name = strings.ToUpper(string(line[:i]))

	if i >= len(line) {
		return p
	}

	// Парсим параметры, если есть ';'.
	if line[i] == ';' {
		i++
		i = parseParamsInto(&p, line, i)
	}

	// Пропускаем ':' и берём значение.
	if i < len(line) && line[i] == ':' {
		i++
	}
	p.Value = string(line[i:])

	return p
}

// parseParamsInto парсит параметры из line начиная с позиции start.
// Возвращает позицию после разбора (обычно — на ':',).
func parseParamsInto(p *model.Property, line []byte, start int) int {
	i := start

	for i < len(line) {
		// Ищем '='.
		eqPos := i
		for eqPos < len(line) && line[eqPos] != '=' {
			eqPos++
		}
		if eqPos >= len(line) {
			return i
		}
		paramName := strings.ToUpper(string(line[i:eqPos]))
		i = eqPos + 1

		// Парсим значения параметра (через запятую, с поддержкой кавычек).
		var values []string
		for {
			val, end := parseParamValue(line, i)
			values = append(values, decodeParamValueRFC6868(val))
			i = end
			if i >= len(line) || line[i] != ',' {
				break
			}
			i++ // пропускаем ','
		}

		p.Params = append(p.Params, model.Param{Name: paramName, Values: values})

		if i >= len(line) || line[i] == ':' {
			return i
		}
		i++ // пропускаем ';'
	}

	return i
}

// parseParamValue парсит одно значение параметра (в кавычках или без).
func parseParamValue(line []byte, start int) (value string, end int) {
	if start < len(line) && line[start] == '"' {
		// Значение в кавычках.
		end = start + 1
		for end < len(line) && line[end] != '"' {
			end++
		}
		value = string(line[start+1 : end])
		if end < len(line) {
			end++ // пропускаем закрывающую кавычку
		}
		return value, end
	}
	// Значение без кавычек — до ',', ';' или ':'.
	end = start
	for end < len(line) && line[end] != ',' && line[end] != ';' && line[end] != ':' {
		end++
	}
	return string(line[start:end]), end
}

// unescapeText де-экранирует TEXT-значение (RFC 5545 §3.3.11).
//
// Правила:
//   - \n или \N → символ новой строки (\n)
//   - \, → запятая
//   - \; → точка с запятой
//   - \\ → обратный слеш
//
// Неизвестные последовательности остаются без изменений.
// Если в строке нет '\' — возвращает оригинал без аллокаций.
func unescapeText(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var sb strings.Builder
	sb.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			sb.WriteByte(s[i])
			continue
		}
		next := s[i+1]
		switch next {
		case 'n', 'N':
			sb.WriteByte('\n')
			i++
		case ',':
			sb.WriteByte(',')
			i++
		case ';':
			sb.WriteByte(';')
			i++
		case '\\':
			sb.WriteByte('\\')
			i++
		default:
			sb.WriteByte('\\')
		}
	}
	return sb.String()
}

// decodeParamValueRFC6868 декодирует ^-escape для параметров (RFC 6868).
// ^n -> перевод строки, ^^ -> ^, ^' -> ".
// Неизвестные последовательности остаются без изменений.
func decodeParamValueRFC6868(value string) string {
	if !strings.Contains(value, "^") {
		return value
	}
	var sb strings.Builder
	sb.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch != '^' || i+1 >= len(value) {
			sb.WriteByte(ch)
			continue
		}
		next := value[i+1]
		switch next {
		case 'n', 'N':
			sb.WriteByte('\n')
			i++
		case '^':
			sb.WriteByte('^')
			i++
		case '\'':
			sb.WriteByte('"')
			i++
		default:
			sb.WriteByte('^')
		}
	}
	return sb.String()
}
