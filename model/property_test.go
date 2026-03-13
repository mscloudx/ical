package model_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"gitverse.ru/cloudcoder/ical/model"
)

// PropertySuite — набор тестов для методов Property.
type PropertySuite struct {
	suite.Suite
}

func TestPropertySuite(t *testing.T) {
	suite.Run(t, new(PropertySuite))
}

func (s *PropertySuite) TestParamValue_ReturnsValueByName() {
	// Arrange
	p := model.Property{
		Name: "DTSTART",
		Params: []model.Param{
			{Name: "TZID", Values: []string{"Europe/Moscow"}},
			{Name: "VALUE", Values: []string{"DATE-TIME"}},
		},
		Value: "20231025T090000",
	}

	// Act
	result := p.ParamValue("TZID")

	// Assert
	s.Equal("Europe/Moscow", result)
}

func (s *PropertySuite) TestParamValue_CaseInsensitive() {
	// Arrange
	p := model.Property{
		Params: []model.Param{
			{Name: "TZID", Values: []string{"America/New_York"}},
		},
	}

	// Act
	result := p.ParamValue("tzid")

	// Assert
	s.Equal("America/New_York", result)
}

func (s *PropertySuite) TestParamValue_ReturnsEmptyWhenParamAbsent() {
	// Arrange
	p := model.Property{
		Name:  "SUMMARY",
		Value: "Test",
	}

	// Act
	result := p.ParamValue("TZID")

	// Assert
	s.Empty(result)
}

func (s *PropertySuite) TestParamValue_ReturnsEmptyWhenValuesEmpty() {
	// Arrange
	p := model.Property{
		Params: []model.Param{
			{Name: "TZID", Values: []string{}},
		},
	}

	// Act
	result := p.ParamValue("TZID")

	// Assert
	s.Empty(result)
}

func (s *PropertySuite) TestHasParam_ReturnsTrueWhenPresent() {
	// Arrange
	p := model.Property{
		Params: []model.Param{
			{Name: "TZID", Values: []string{"Europe/Moscow"}},
		},
	}

	// Act & Assert
	s.True(p.HasParam("TZID"))
	s.True(p.HasParam("tzid"))
}

func (s *PropertySuite) TestHasParam_ReturnsFalseWhenAbsent() {
	// Arrange
	p := model.Property{
		Name:  "SUMMARY",
		Value: "Test",
	}

	// Act & Assert
	s.False(p.HasParam("TZID"))
}
