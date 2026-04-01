// Package model содержит модели данных iCalendar (RFC 5545).
//
// Пакет предоставляет типизированные Go-структуры для представления
// компонентов, свойств и значений стандарта iCalendar.
// Все перечисления реализованы через iota для memory-эффективности;
// конвертация в строки RFC выполняется на уровне парсера/билдера.
package model

import "time"

// --------------------------------------------------------------------------
// Status — статус компонента (STATUS).
// RFC 5545 §3.8.1.11.
// --------------------------------------------------------------------------

// Status определяет текущий статус компонента.
type Status int

const (
	// StatusUnspecified — статус не задан (нулевое значение).
	StatusUnspecified Status = iota
	// StatusTentative — предварительный (VEVENT).
	StatusTentative
	// StatusConfirmed — подтверждённый (VEVENT).
	StatusConfirmed
	// StatusCancelled — отменённый (VEVENT, VTODO, VJOURNAL).
	StatusCancelled
	// StatusCompleted — выполнен (VTODO).
	StatusCompleted
	// StatusNeedsAction — требует действия (VTODO).
	StatusNeedsAction
	// StatusInProcess — в процессе выполнения (VTODO).
	StatusInProcess
	// StatusDraft — черновик (VJOURNAL).
	StatusDraft
	// StatusFinal — финальная версия (VJOURNAL).
	StatusFinal
)

// --------------------------------------------------------------------------
// Transparency — прозрачность времени (TRANSP).
// RFC 5545 §3.8.2.7.
// --------------------------------------------------------------------------

// Transparency определяет, влияет ли событие на поиск свободного времени.
type Transparency int

const (
	// TranspUnspecified — прозрачность не задана (нулевое значение).
	TranspUnspecified Transparency = iota
	// TranspOpaque — событие перекрывает время (занят).
	TranspOpaque
	// TranspTransparent — событие не перекрывает время (свободен).
	TranspTransparent
)

// --------------------------------------------------------------------------
// Classification — уровень доступа (CLASS).
// RFC 5545 §3.8.1.3.
// --------------------------------------------------------------------------

// Classification определяет уровень видимости компонента.
type Classification int

const (
	// ClassUnspecified — уровень доступа не задан (нулевое значение).
	ClassUnspecified Classification = iota
	// ClassPublic — публичный.
	ClassPublic
	// ClassPrivate — приватный.
	ClassPrivate
	// ClassConfidential — конфиденциальный.
	ClassConfidential
)

// --------------------------------------------------------------------------
// ParticipationStatus — статус участия (PARTSTAT).
// RFC 5545 §3.2.12.
// --------------------------------------------------------------------------

// ParticipationStatus определяет статус участия в событии.
type ParticipationStatus int

const (
	// PartStatUnspecified — статус участия не задан (нулевое значение).
	PartStatUnspecified ParticipationStatus = iota
	// PartStatNeedsAction — ожидает ответа.
	PartStatNeedsAction
	// PartStatAccepted — принял приглашение.
	PartStatAccepted
	// PartStatDeclined — отклонил приглашение.
	PartStatDeclined
	// PartStatTentative — предварительно принял.
	PartStatTentative
	// PartStatDelegated — делегировал другому участнику.
	PartStatDelegated
)

// --------------------------------------------------------------------------
// Role — роль участника (ROLE).
// RFC 5545 §3.2.16.
// --------------------------------------------------------------------------

// Role определяет роль участника в событии.
type Role int

const (
	// RoleUnspecified — роль не задана (нулевое значение).
	RoleUnspecified Role = iota
	// RoleChair — председатель.
	RoleChair
	// RoleReqParticipant — обязательный участник.
	RoleReqParticipant
	// RoleOptParticipant — необязательный участник.
	RoleOptParticipant
	// RoleNonParticipant — не участник (информирование).
	RoleNonParticipant
)

// --------------------------------------------------------------------------
// AlarmAction — действие напоминания (ACTION).
// RFC 5545 §3.8.6.1.
// --------------------------------------------------------------------------

// AlarmAction определяет тип действия при напоминании.
type AlarmAction int

const (
	// ActionUnspecified — действие не задано (нулевое значение).
	ActionUnspecified AlarmAction = iota
	// ActionAudio — звуковое напоминание.
	ActionAudio
	// ActionDisplay — визуальное напоминание.
	ActionDisplay
	// ActionEmail — напоминание по email.
	ActionEmail
)

// --------------------------------------------------------------------------
// RelationshipType — тип связи компонентов (RELTYPE).
// RFC 5545 §3.2.15.
// --------------------------------------------------------------------------

// RelationshipType определяет тип связи между компонентами.
type RelationshipType int

const (
	// RelTypeUnspecified — тип связи не задан (нулевое значение).
	RelTypeUnspecified RelationshipType = iota
	// RelTypeParent — родительский компонент.
	RelTypeParent
	// RelTypeChild — дочерний компонент.
	RelTypeChild
	// RelTypeSibling — компонент того же уровня.
	RelTypeSibling
	// RelTypeFinishToStart — связь окончание→начало.
	RelTypeFinishToStart
	// RelTypeFinishToFinish — связь окончание→окончание.
	RelTypeFinishToFinish
	// RelTypeStartToFinish — связь начало→окончание.
	RelTypeStartToFinish
	// RelTypeStartToStart — связь начало→начало.
	RelTypeStartToStart
	// RelTypeFirst — первый в цепочке.
	RelTypeFirst
	// RelTypeNext — следующий в цепочке.
	RelTypeNext
	// RelTypeDependsOn — зависит от другого компонента.
	RelTypeDependsOn
	// RelTypeRefID — связь через REFID.
	RelTypeRefID
	// RelTypeConcept — связь через CONCEPT.
	RelTypeConcept
	// RelTypeSnooze — связь "snooze" между VALARM (RFC 9074).
	RelTypeSnooze
)

// --------------------------------------------------------------------------
// FreeBusyType — тип периода занятости (FBTYPE).
// RFC 5545 §3.2.9.
// --------------------------------------------------------------------------

// FreeBusyType определяет тип периода свободного/занятого времени.
type FreeBusyType int

const (
	// FBTypeUnspecified — тип не задан (нулевое значение).
	FBTypeUnspecified FreeBusyType = iota
	// FBTypeFree — свободен.
	FBTypeFree
	// FBTypeBusy — занят.
	FBTypeBusy
	// FBTypeBusyUnavailable — занят, недоступен.
	FBTypeBusyUnavailable
	// FBTypeBusyTentative — занят предварительно.
	FBTypeBusyTentative
)

// --------------------------------------------------------------------------
// Method — метод iTIP (RFC 5546 §3.2).
// RFC 5545 §3.7.2.
// --------------------------------------------------------------------------

// Method определяет метод iTIP для iCalendar-объекта.
type Method int

const (
	// MethodUnspecified — метод не задан (нулевое значение).
	MethodUnspecified Method = iota
	// MethodPublish — публикация компонента без ответа.
	MethodPublish
	// MethodRequest — запрос участия или обновление компонента.
	MethodRequest
	// MethodReply — ответ на REQUEST от участника.
	MethodReply
	// MethodAdd — добавление экземпляра к повторяющемуся компоненту.
	MethodAdd
	// MethodCancel — отмена компонента.
	MethodCancel
	// MethodRefresh — запрос актуальной версии компонента.
	MethodRefresh
	// MethodCounter — встречное предложение по параметрам компонента.
	MethodCounter
	// MethodDeclineCounter — отклонение встречного предложения.
	MethodDeclineCounter
)

// String возвращает RFC 5546 строку для метода iTIP.
func (m Method) String() string {
	switch m {
	case MethodPublish:
		return "PUBLISH"
	case MethodRequest:
		return "REQUEST"
	case MethodReply:
		return "REPLY"
	case MethodAdd:
		return "ADD"
	case MethodCancel:
		return "CANCEL"
	case MethodRefresh:
		return "REFRESH"
	case MethodCounter:
		return "COUNTER"
	case MethodDeclineCounter:
		return "DECLINECOUNTER"
	default:
		return ""
	}
}

// --------------------------------------------------------------------------
// ScheduleAgent — агент планирования (RFC 6638 §7.1).
// --------------------------------------------------------------------------

// ScheduleAgent определяет, кто выполняет iTIP-обработку для участника.
type ScheduleAgent int

const (
	// ScheduleAgentUnspecified — агент не задан (нулевое значение, сервер обрабатывает по умолчанию).
	ScheduleAgentUnspecified ScheduleAgent = iota
	// ScheduleAgentServer — сервер CalDAV выполняет iTIP-рассылку.
	ScheduleAgentServer
	// ScheduleAgentClient — клиент берёт iTIP-обработку на себя.
	ScheduleAgentClient
	// ScheduleAgentNone — iTIP-обработка отключена.
	ScheduleAgentNone
)

// String возвращает RFC 6638 строку для агента планирования.
func (sa ScheduleAgent) String() string {
	switch sa {
	case ScheduleAgentServer:
		return "SERVER"
	case ScheduleAgentClient:
		return "CLIENT"
	case ScheduleAgentNone:
		return "NONE"
	default:
		return ""
	}
}

// --------------------------------------------------------------------------
// RequestStatus — статус обработки iTIP-запроса (RFC 5546 §3.6).
// RFC 5545 §3.8.8.3.
// --------------------------------------------------------------------------

// RequestStatus содержит статусный код и описание результата iTIP-обработки.
//
// Формат значения: <код>;<описание>[;<дополнительные данные>]
// Примеры кодов:
//
//	2.0  — Success
//	3.7  — Invalid calendar user
//	4.1  — Event conflict; date not suitable
type RequestStatus struct {
	// Code — трёхзначный код статуса (например, "2.0", "3.7").
	Code string
	// Description — текстовое описание статуса.
	Description string
	// ExtraData — дополнительные данные (опционально, например спорная часть запроса).
	ExtraData string
}

// --------------------------------------------------------------------------
// Переиспользуемые структуры
// --------------------------------------------------------------------------

// Attendee представляет участника или организатора (ATTENDEE / ORGANIZER).
// RFC 5545 §3.8.4.1, §3.8.4.3.
type Attendee struct {
	// Address — cal-address участника (например, "mailto:user@example.com").
	Address string
	// Name — отображаемое имя (параметр CN).
	Name string
	// Role — роль участника.
	Role Role
	// Status — статус участия (PARTSTAT).
	Status ParticipationStatus
	// RSVP — запрос подтверждения участия.
	RSVP bool
	// ScheduleAgent — кто выполняет iTIP-рассылку (RFC 6638 §7.1).
	ScheduleAgent ScheduleAgent
	// ScheduleForceSend — принудительная отправка iTIP-сообщения (RFC 6638 §7.2).
	// Значения: "REQUEST", "REPLY" или пустая строка.
	ScheduleForceSend string
	// ScheduleStatus — статус планирования от сервера (RFC 6638 §7.3).
	// Список кодов, разделённых запятой (например, "1.2" или "3.7").
	ScheduleStatus []string
	// Params — дополнительные параметры, не покрытые полями выше.
	Params []Param
}

// Relation представляет связь RELATED-TO между компонентами.
// RFC 5545 §3.8.4.5.
type Relation struct {
	// Type — тип связи (PARENT, CHILD, SIBLING).
	Type RelationshipType
	// UID — уникальный идентификатор связанного компонента.
	UID string
	// Gap — смещение/зазор между компонентами (GAP).
	Gap *time.Duration
}

// Link представляет свойство LINK (RFC 9253).
type Link struct {
	// Value — значение свойства LINK.
	Value string
	// ValueType — тип значения (VALUE=URI/UID/XML-REFERENCE).
	ValueType string
	// Rel — тип связи (LINKREL).
	Rel string
	// FmtType — MIME-тип (FMTTYPE).
	FmtType string
	// Label — текстовая метка (LABEL).
	Label string
	// Language — язык метки (LANGUAGE).
	Language string
	// Params — дополнительные параметры LINK, не покрытые полями выше.
	Params []Param
}

// Period представляет временной период (PERIOD value type).
// RFC 5545 §3.3.9.
//
// Период задаётся либо парой Start/End, либо Start + Duration.
// Если Duration > 0 и End нулевой — используется Duration.
type Period struct {
	// Start — начало периода.
	Start time.Time
	// End — конец периода (если задан явно).
	End time.Time
	// Duration — длительность периода (альтернатива End).
	Duration time.Duration
}

// --------------------------------------------------------------------------
// Методы String() для enum-типов — возвращают RFC 5545 строки.
// --------------------------------------------------------------------------

// String возвращает RFC 5545 строку для статуса.
func (s Status) String() string {
	switch s {
	case StatusTentative:
		return "TENTATIVE"
	case StatusConfirmed:
		return "CONFIRMED"
	case StatusCancelled:
		return "CANCELLED"
	case StatusCompleted:
		return "COMPLETED"
	case StatusNeedsAction:
		return "NEEDS-ACTION"
	case StatusInProcess:
		return "IN-PROCESS"
	case StatusDraft:
		return "DRAFT"
	case StatusFinal:
		return "FINAL"
	default:
		return ""
	}
}

// String возвращает RFC 5545 строку для прозрачности.
func (t Transparency) String() string {
	switch t {
	case TranspOpaque:
		return "OPAQUE"
	case TranspTransparent:
		return "TRANSPARENT"
	default:
		return ""
	}
}

// String возвращает RFC 5545 строку для уровня доступа.
func (c Classification) String() string {
	switch c {
	case ClassPublic:
		return "PUBLIC"
	case ClassPrivate:
		return "PRIVATE"
	case ClassConfidential:
		return "CONFIDENTIAL"
	default:
		return ""
	}
}

// String возвращает RFC 5545 строку для статуса участия.
func (ps ParticipationStatus) String() string {
	switch ps {
	case PartStatNeedsAction:
		return "NEEDS-ACTION"
	case PartStatAccepted:
		return "ACCEPTED"
	case PartStatDeclined:
		return "DECLINED"
	case PartStatTentative:
		return "TENTATIVE"
	case PartStatDelegated:
		return "DELEGATED"
	default:
		return ""
	}
}

// String возвращает RFC 5545 строку для роли участника.
func (r Role) String() string {
	switch r {
	case RoleChair:
		return "CHAIR"
	case RoleReqParticipant:
		return "REQ-PARTICIPANT"
	case RoleOptParticipant:
		return "OPT-PARTICIPANT"
	case RoleNonParticipant:
		return "NON-PARTICIPANT"
	default:
		return ""
	}
}

// String возвращает RFC 5545 строку для действия напоминания.
func (a AlarmAction) String() string {
	switch a {
	case ActionAudio:
		return "AUDIO"
	case ActionDisplay:
		return "DISPLAY"
	case ActionEmail:
		return "EMAIL"
	default:
		return ""
	}
}

// String возвращает RFC 5545 строку для типа связи.
func (rt RelationshipType) String() string {
	switch rt {
	case RelTypeParent:
		return "PARENT"
	case RelTypeChild:
		return "CHILD"
	case RelTypeSibling:
		return "SIBLING"
	case RelTypeFinishToStart:
		return "FINISHTOSTART"
	case RelTypeFinishToFinish:
		return "FINISHTOFINISH"
	case RelTypeStartToFinish:
		return "STARTTOFINISH"
	case RelTypeStartToStart:
		return "STARTTOSTART"
	case RelTypeFirst:
		return "FIRST"
	case RelTypeNext:
		return "NEXT"
	case RelTypeDependsOn:
		return "DEPENDS-ON"
	case RelTypeRefID:
		return "REFID"
	case RelTypeConcept:
		return "CONCEPT"
	case RelTypeSnooze:
		return "SNOOZE"
	default:
		return ""
	}
}

// Proximity определяет тип proximity-триггера для VALARM (RFC 9074).
// Допускает как стандартные значения, так и iana-token/x-name.
type Proximity string

const (
	// ProximityArrive — триггер при прибытии.
	ProximityArrive Proximity = "ARRIVE"
	// ProximityDepart — триггер при убытии.
	ProximityDepart Proximity = "DEPART"
	// ProximityConnect — триггер при подключении.
	ProximityConnect Proximity = "CONNECT"
	// ProximityDisconnect — триггер при отключении.
	ProximityDisconnect Proximity = "DISCONNECT"
)

// String возвращает строковое значение proximity.
func (p Proximity) String() string {
	return string(p)
}

// String возвращает RFC 5545 строку для типа занятости.
func (fb FreeBusyType) String() string {
	switch fb {
	case FBTypeFree:
		return "FREE"
	case FBTypeBusy:
		return "BUSY"
	case FBTypeBusyUnavailable:
		return "BUSY-UNAVAILABLE"
	case FBTypeBusyTentative:
		return "BUSY-TENTATIVE"
	default:
		return ""
	}
}

// --------------------------------------------------------------------------
// Вспомогательные функции
// --------------------------------------------------------------------------

// IsDateOnly проверяет, представляет ли time.Time значение только даты
// (VALUE=DATE), без временной части.
func IsDateOnly(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
}
