package model

// Attachment представляет свойство ATTACH — вложение или ссылку на ресурс.
//
// Допускается ровно одно из двух: URI или встроенные двоичные данные (Data).
// При наличии Data — при сериализации кодируется в BASE64 (ENCODING=BASE64;VALUE=BINARY).
// RFC 5545 §3.8.1.1.
type Attachment struct {
	// URI — ссылка на прикреплённый ресурс.
	// Взаимоисключающее с Data.
	URI string
	// Data — встроенные двоичные данные.
	// При сериализации кодируются в BASE64; взаимоисключающее с URI.
	Data []byte
	// MIMEType — MIME-тип содержимого (параметр FMTTYPE).
	// Например: "application/pdf", "image/png".
	MIMEType string
	// Params — дополнительные параметры свойства ATTACH (например, FILENAME, MANAGED-ID, SIZE).
	// Хранит все параметры кроме FMTTYPE и ENCODING/VALUE.
	Params []Param
}
