package builder

import (
	"time"

	"github.com/mscloudx/ical/model"
)

// AlarmOption — функция настройки Alarm.
type AlarmOption func(*model.Alarm)

// NewDisplayAlarm создаёт визуальное напоминание (ACTION:DISPLAY).
func NewDisplayAlarm(trigger model.Trigger, description string, opts ...AlarmOption) model.Alarm {
	a := model.Alarm{
		Action:      model.ActionDisplay,
		Trigger:     trigger,
		Description: description,
	}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

// NewAudioAlarm создаёт звуковое напоминание (ACTION:AUDIO).
func NewAudioAlarm(trigger model.Trigger, opts ...AlarmOption) model.Alarm {
	a := model.Alarm{
		Action:  model.ActionAudio,
		Trigger: trigger,
	}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

// NewEmailAlarm создаёт email-напоминание (ACTION:EMAIL).
func NewEmailAlarm(trigger model.Trigger, summary, description string, attendees []model.Attendee, opts ...AlarmOption) model.Alarm {
	a := model.Alarm{
		Action:      model.ActionEmail,
		Trigger:     trigger,
		Summary:     summary,
		Description: description,
		Attendees:   attendees,
	}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

// TriggerBefore возвращает триггер, срабатывающий за d до начала события.
// Принимает положительный Duration, который инвертируется.
func TriggerBefore(d time.Duration) model.Trigger {
	dur := -d
	return model.Trigger{Duration: &dur}
}

// TriggerAt возвращает триггер, срабатывающий в абсолютное время.
func TriggerAt(t time.Time) model.Trigger {
	return model.Trigger{DateTime: &t}
}

func WithAlarmDuration(d time.Duration) AlarmOption {
	return func(a *model.Alarm) {
		a.Duration = &d
	}
}

func WithAlarmRepeat(r int) AlarmOption {
	return func(a *model.Alarm) {
		a.Repeat = r
	}
}

func WithAlarmUID(uid string) AlarmOption {
	return func(a *model.Alarm) {
		a.UID = uid
	}
}

func WithAlarmAcknowledged(t time.Time) AlarmOption {
	return func(a *model.Alarm) {
		a.Acknowledged = &t
	}
}

func WithAlarmProximity(p model.Proximity) AlarmOption {
	return func(a *model.Alarm) {
		a.Proximity = p
	}
}

func WithAlarmRelated(rel model.Relation) AlarmOption {
	return func(a *model.Alarm) {
		a.Related = append(a.Related, rel)
	}
}

func WithAlarmLocation(loc *model.LocationComponent) AlarmOption {
	return func(a *model.Alarm) {
		if loc == nil {
			return
		}
		a.Locations = append(a.Locations, *loc)
	}
}

func WithAlarmXProp(name, value string, params ...model.Param) AlarmOption {
	return func(a *model.Alarm) {
		a.XProps = append(a.XProps, model.Property{Name: name, Value: value, Params: params})
	}
}

// WithAlarmAttachment добавляет вложение ATTACH к напоминанию (ACTION:EMAIL, RFC 5545 §3.8.1.1).
func WithAlarmAttachment(att ...model.Attachment) AlarmOption {
	return func(a *model.Alarm) {
		a.Attach = append(a.Attach, att...)
	}
}
