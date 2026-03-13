package model

import "strings"

// Param — параметр свойства iCalendar.
//
// Параметры идут после имени свойства через ';' и уточняют его значение.
// Пример: TZID=Europe/Moscow, ROLE=REQ-PARTICIPANT.
// RFC 5545 §3.2.
type Param struct {
	// Name — имя параметра в верхнем регистре (например, "TZID", "CN").
	Name string
	// Values — одно или несколько значений параметра.
	Values []string
}

// Property представляет одну content-line из iCalendar.
//
// Content-line имеет формат:
//
//	ИмяСвойства[;Параметр1=Значение1;...]:ЗначениеСвойства
//
// RFC 5545 §3.1.
type Property struct {
	// Name — имя свойства в верхнем регистре (например, "DTSTART", "SUMMARY").
	Name string
	// Params — список параметров свойства.
	Params []Param
	// Value — значение свойства (строка после ':').
	Value string
}

// ParamValue возвращает первое значение параметра с указанным именем.
// Если параметр не найден или не имеет значений — возвращает пустую строку.
// Поиск регистронезависимый.
func (p Property) ParamValue(name string) string {
	for _, param := range p.Params {
		if strings.EqualFold(param.Name, name) && len(param.Values) > 0 {
			return param.Values[0]
		}
	}
	return ""
}

// HasParam проверяет наличие параметра с указанным именем.
// Поиск регистронезависимый.
func (p Property) HasParam(name string) bool {
	for _, param := range p.Params {
		if strings.EqualFold(param.Name, name) {
			return true
		}
	}
	return false
}
