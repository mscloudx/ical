package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gitverse.ru/cloudcoder/ical/model"
)

// TypesSuite — набор тестов для вспомогательных функций и типов.
type TypesSuite struct {
	suite.Suite
}

func TestTypesSuite(t *testing.T) {
	suite.Run(t, new(TypesSuite))
}

func (s *TypesSuite) TestIsDateOnly_ReturnsTrueForDateOnly() {
	// Arrange
	t := time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC)

	// Act
	result := model.IsDateOnly(t)

	// Assert
	s.True(result)
}

func (s *TypesSuite) TestIsDateOnly_ReturnsFalseForDateTime() {
	// Arrange
	t := time.Date(2023, 10, 25, 9, 0, 0, 0, time.UTC)

	// Act
	result := model.IsDateOnly(t)

	// Assert
	s.False(result)
}

func (s *TypesSuite) TestIsDateOnly_ReturnsFalseForNanoseconds() {
	// Arrange
	t := time.Date(2023, 10, 25, 0, 0, 0, 1, time.UTC)

	// Act
	result := model.IsDateOnly(t)

	// Assert
	s.False(result)
}

func (s *TypesSuite) TestEnums_ZeroValueIsUnspecified() {
	// Arrange & Act & Assert
	s.Equal(model.StatusUnspecified, model.Status(0))
	s.Equal(model.TranspUnspecified, model.Transparency(0))
	s.Equal(model.ClassUnspecified, model.Classification(0))
	s.Equal(model.PartStatUnspecified, model.ParticipationStatus(0))
	s.Equal(model.RoleUnspecified, model.Role(0))
	s.Equal(model.ActionUnspecified, model.AlarmAction(0))
	s.Equal(model.RelTypeUnspecified, model.RelationshipType(0))
	s.Equal(model.FBTypeUnspecified, model.FreeBusyType(0))
}

func (s *TypesSuite) TestEvent_ZeroValueIsSafe() {
	// Arrange & Act
	var e model.Event

	// Assert
	s.Empty(e.UID)
	s.True(e.DTStamp.IsZero())
	s.True(e.DTStart.IsZero())
	s.Nil(e.DTEnd)
	s.Nil(e.Duration)
	s.False(e.AllDay)
	s.Equal(model.StatusUnspecified, e.Status)
	s.Nil(e.Organizer)
	s.Nil(e.Attendees)
	s.Nil(e.RRules)
	s.Nil(e.Alarms)
	s.Nil(e.XProps)
}
