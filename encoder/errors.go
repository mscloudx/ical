package encoder

import "github.com/pkg/errors"

// Sentinel-ошибки энкодера iCalendar.
var (
	// ErrNilCalendar возвращается, если передан nil Calendar.
	ErrNilCalendar = errors.New("calendar is nil")
)
