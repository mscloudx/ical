package builder

import "github.com/pkg/errors"

// Sentinel-ошибки builder iCalendar.
var (
	// ErrEmptyProdID возвращается при создании Calendar с пустым PRODID.
	// PRODID — обязательное свойство VCALENDAR (RFC 5545 §3.7.3).
	ErrEmptyProdID = errors.New("PRODID is required")

	// ErrEmptyUID возвращается при создании компонента с пустым UID.
	// UID — обязательное свойство для VEVENT, VTODO, VJOURNAL (RFC 5545 §3.8.4.7).
	ErrEmptyUID = errors.New("UID is required")
)
