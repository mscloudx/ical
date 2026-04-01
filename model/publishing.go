package model

import "time"

// StructuredData представляет свойство STRUCTURED-DATA (RFC 9073).
type StructuredData struct {
	Value     string // само значение (может быть URI или текстовые данные)
	ValueType string // VALUE=URI (по умолчанию TEXT)
	FmtType   string // FMTTYPE (например, "application/json")
	Schema    string // SCHEMA (URI схемы)
}

// StyledDescription представляет свойство STYLED-DESCRIPTION (RFC 9073).
type StyledDescription struct {
	Value     string
	ValueType string // VALUE=URI или TEXT
	FmtType   string // FMTTYPE (например, "text/html")
	Language  string // LANGUAGE
}

// Participant представляет компонент PARTICIPANT (RFC 9073).
type Participant struct {
	UID             string
	ParticipantType string // KIND (например, INDIVIDUAL, GROUP, RESOURCE, ROOM)
	CalendarAddress string // CALENDAR-ADDRESS (URI)

	// Данные внутри PARTICIPANT
	Created      *time.Time
	Description  string
	DTStamp      time.Time
	LastModified *time.Time
	Sequence     int
	Summary      string
	URL          string
	Categories   []string
	Contacts     []string

	Location       string // встроенное свойство LOCATION
	StructuredData []StructuredData

	// Вложенные компоненты в PARTICIPANT
	Locations []LocationComponent
	Resources []ResourceComponent

	// Concepts — семантические категории (CONCEPT, RFC 9253).
	Concepts []string
	// Links — связи с внешними ресурсами (LINK, RFC 9253).
	Links []Link
	// RefIDs — идентификаторы связей (REFID, RFC 9253).
	RefIDs []string

	// XProps — нестандартные свойства.
	XProps []Property
	// IanaProps — зарегистрированные свойства IANA.
	IanaProps []Property
}

// LocationComponent представляет компонент LOCATION (RFC 9073).
type LocationComponent struct {
	UID         string
	Name        string
	Description string
	Geo         string   // GEO (опционально можно использовать структуру)
	Type        []string // LOCATION-TYPE

	Created      *time.Time
	DTStamp      time.Time
	LastModified *time.Time
	Sequence     int
	URL          string

	StructuredData []StructuredData

	// Concepts — семантические категории (CONCEPT, RFC 9253).
	Concepts []string
	// Links — связи с внешними ресурсами (LINK, RFC 9253).
	Links []Link
	// RefIDs — идентификаторы связей (REFID, RFC 9253).
	RefIDs []string

	// XProps — нестандартные свойства.
	XProps []Property
	// IanaProps — зарегистрированные свойства IANA.
	IanaProps []Property
}

// ResourceComponent представляет компонент RESOURCE (RFC 9073).
type ResourceComponent struct {
	UID          string
	Name         string
	Description  string
	ResourceType []string // RESOURCE-TYPE

	Created      *time.Time
	DTStamp      time.Time
	LastModified *time.Time
	Sequence     int
	URL          string

	StructuredData []StructuredData

	// Concepts — семантические категории (CONCEPT, RFC 9253).
	Concepts []string
	// Links — связи с внешними ресурсами (LINK, RFC 9253).
	Links []Link
	// RefIDs — идентификаторы связей (REFID, RFC 9253).
	RefIDs []string

	// XProps — нестандартные свойства.
	XProps []Property
	// IanaProps — зарегистрированные свойства IANA.
	IanaProps []Property
}
