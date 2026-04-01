package parser

import (
	"strconv"
	"strings"

	"github.com/mscloudx/ical/model"
	"github.com/pkg/errors"
)

const (
	componentLocation    = "LOCATION"
	componentResource    = "RESOURCE"
	componentAlarm       = "VALARM"
	componentParticipant = "PARTICIPANT"

	propKind              = "KIND"
	propParticipantType   = "PARTICIPANT-TYPE"
	propCalendarAddress   = "CALENDAR-ADDRESS"
	propCategories        = "CATEGORIES"
	propContact           = "CONTACT"
	propLocation          = "LOCATION"
	propStructuredData    = "STRUCTURED-DATA"
	propStyledDescription = "STYLED-DESCRIPTION"
	propConcept           = "CONCEPT"
	propRefID             = "REFID"
	propLink              = "LINK"

	propName         = "NAME"
	propGeo          = "GEO"
	propLocationType = "LOCATION-TYPE"
	propResourceType = "RESOURCE-TYPE"

	paramValue    = "VALUE"
	paramFmtType  = "FMTTYPE"
	paramSchema   = "SCHEMA"
	paramLanguage = "LANGUAGE"
	paramLinkRel  = "LINKREL"
	paramLabel    = "LABEL"
)

// buildParticipant строит Participant из rawComponent.
func buildParticipant(raw *rawComponent) (model.Participant, error) {
	var part model.Participant

	for _, p := range raw.props {
		if err := setParticipantProp(&part, p); err != nil {
			return part, err
		}
	}

	for _, child := range raw.children {
		switch child.name {
		case componentLocation:
			loc, err := buildLocationComponent(child)
			if err != nil {
				return part, errors.Wrap(err, componentLocation)
			}
			part.Locations = append(part.Locations, loc)
		case componentResource:
			res, err := buildResourceComponent(child)
			if err != nil {
				return part, errors.Wrap(err, componentResource)
			}
			part.Resources = append(part.Resources, res)
		}
	}

	if part.UID == "" {
		return part, errors.Wrap(ErrInvalidValue, "PARTICIPANT: missing UID")
	}
	if part.ParticipantType == "" {
		return part, errors.Wrap(ErrInvalidValue, "PARTICIPANT: missing PARTICIPANT-TYPE")
	}

	return part, nil
}

func setParticipantProp(part *model.Participant, p model.Property) error {
	switch strings.ToUpper(p.Name) {
	case propUID:
		part.UID = p.Value
	case propKind, propParticipantType:
		part.ParticipantType = p.Value
	case propCalendarAddress:
		part.CalendarAddress = p.Value
	case propCREATED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		part.Created = &t
	case propDESCRIPTION:
		part.Description = unescapeText(p.Value)
	case propDTSTAMP:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		part.DTStamp = t
	case propLASTMODIFIED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		part.LastModified = &t
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "SEQUENCE=%s", p.Value)
		}
		part.Sequence = n
	case propSUMMARY:
		part.Summary = unescapeText(p.Value)
	case propURL:
		part.URL = p.Value
	case propCategories:
		cats := strings.Split(p.Value, ",")
		for i := range cats {
			cats[i] = strings.TrimSpace(cats[i])
		}
		part.Categories = append(part.Categories, cats...)
	case propContact:
		part.Contacts = append(part.Contacts, p.Value)
	case propLocation: // Свойство внутри PARTICIPANT
		part.Location = unescapeText(p.Value)
	case propStructuredData:
		part.StructuredData = append(part.StructuredData, parseStructuredData(p))
	case propConcept:
		part.Concepts = append(part.Concepts, p.Value)
	case propRefID:
		part.RefIDs = append(part.RefIDs, p.Value)
	case propLink:
		part.Links = append(part.Links, parseLink(p))
	default:
		switch {
		case isXProp(p.Name):
			part.XProps = append(part.XProps, p)
		case isIanaProp(p.Name):
			part.IanaProps = append(part.IanaProps, p)
		}
	}
	return nil
}

// buildLocationComponent строит LocationComponent из rawComponent.
func buildLocationComponent(raw *rawComponent) (model.LocationComponent, error) {
	var loc model.LocationComponent

	for _, p := range raw.props {
		if err := setLocationComponentProp(&loc, p); err != nil {
			return loc, err
		}
	}

	if loc.UID == "" {
		return loc, errors.Wrap(ErrInvalidValue, "LOCATION: missing UID")
	}

	return loc, nil
}

func setLocationComponentProp(loc *model.LocationComponent, p model.Property) error {
	switch strings.ToUpper(p.Name) {
	case propUID:
		loc.UID = p.Value
	case propName:
		loc.Name = unescapeText(p.Value)
	case propDESCRIPTION:
		loc.Description = unescapeText(p.Value)
	case propGeo:
		loc.Geo = p.Value
	case propLocationType:
		types := strings.Split(p.Value, ",")
		for i := range types {
			types[i] = strings.TrimSpace(types[i])
		}
		loc.Type = append(loc.Type, types...)
	case propCREATED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		loc.Created = &t
	case propDTSTAMP:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		loc.DTStamp = t
	case propLASTMODIFIED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		loc.LastModified = &t
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "SEQUENCE=%s", p.Value)
		}
		loc.Sequence = n
	case propURL:
		loc.URL = p.Value
	case propStructuredData:
		loc.StructuredData = append(loc.StructuredData, parseStructuredData(p))
	case propConcept:
		loc.Concepts = append(loc.Concepts, p.Value)
	case propRefID:
		loc.RefIDs = append(loc.RefIDs, p.Value)
	case propLink:
		loc.Links = append(loc.Links, parseLink(p))
	default:
		switch {
		case isXProp(p.Name):
			loc.XProps = append(loc.XProps, p)
		case isIanaProp(p.Name):
			loc.IanaProps = append(loc.IanaProps, p)
		}
	}
	return nil
}

// buildResourceComponent строит ResourceComponent из rawComponent.
func buildResourceComponent(raw *rawComponent) (model.ResourceComponent, error) {
	var res model.ResourceComponent

	for _, p := range raw.props {
		if err := setResourceComponentProp(&res, p); err != nil {
			return res, err
		}
	}

	if res.UID == "" {
		return res, errors.Wrap(ErrInvalidValue, "RESOURCE: missing UID")
	}

	return res, nil
}

func setResourceComponentProp(res *model.ResourceComponent, p model.Property) error {
	switch strings.ToUpper(p.Name) {
	case propUID:
		res.UID = p.Value
	case propName:
		res.Name = unescapeText(p.Value)
	case propDESCRIPTION:
		res.Description = unescapeText(p.Value)
	case propResourceType:
		types := strings.Split(p.Value, ",")
		for i := range types {
			types[i] = strings.TrimSpace(types[i])
		}
		res.ResourceType = append(res.ResourceType, types...)
	case propCREATED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propCREATED)
		}
		res.Created = &t
	case propDTSTAMP:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propDTSTAMP)
		}
		res.DTStamp = t
	case propLASTMODIFIED:
		t, _, err := parseDateTime(p.Value, nil)
		if err != nil {
			return errors.Wrap(err, propLASTMODIFIED)
		}
		res.LastModified = &t
	case propSEQUENCE:
		n, err := strconv.Atoi(p.Value)
		if err != nil {
			return errors.Wrapf(ErrInvalidValue, "SEQUENCE=%s", p.Value)
		}
		res.Sequence = n
	case propURL:
		res.URL = p.Value
	case propStructuredData:
		res.StructuredData = append(res.StructuredData, parseStructuredData(p))
	case propConcept:
		res.Concepts = append(res.Concepts, p.Value)
	case propRefID:
		res.RefIDs = append(res.RefIDs, p.Value)
	case propLink:
		res.Links = append(res.Links, parseLink(p))
	default:
		switch {
		case isXProp(p.Name):
			res.XProps = append(res.XProps, p)
		case isIanaProp(p.Name):
			res.IanaProps = append(res.IanaProps, p)
		}
	}
	return nil
}

// parseStructuredData парсит свойство STRUCTURED-DATA.
func parseStructuredData(p model.Property) model.StructuredData {
	sd := model.StructuredData{
		Value: p.Value,
	}
	for _, param := range p.Params {
		switch param.Name {
		case paramValue:
			if len(param.Values) > 0 {
				sd.ValueType = param.Values[0]
			}
		case paramFmtType:
			if len(param.Values) > 0 {
				sd.FmtType = param.Values[0]
			}
		case paramSchema:
			if len(param.Values) > 0 {
				sd.Schema = param.Values[0]
			}
		}
	}
	return sd
}

// parseStyledDescription парсит свойство STYLED-DESCRIPTION.
func parseStyledDescription(p model.Property) model.StyledDescription {
	sd := model.StyledDescription{
		Value: p.Value,
	}
	for _, param := range p.Params {
		switch param.Name {
		case paramValue:
			if len(param.Values) > 0 {
				sd.ValueType = param.Values[0]
			}
		case paramFmtType:
			if len(param.Values) > 0 {
				sd.FmtType = param.Values[0]
			}
		case paramLanguage:
			if len(param.Values) > 0 {
				sd.Language = param.Values[0]
			}
		}
	}
	return sd
}

// parseLink парсит свойство LINK.
func parseLink(p model.Property) model.Link {
	link := model.Link{
		Value: p.Value,
	}
	for _, param := range p.Params {
		if len(param.Values) == 0 {
			continue
		}
		switch param.Name {
		case paramValue:
			if len(param.Values) == 1 {
				link.ValueType = param.Values[0]
			} else {
				link.Params = append(link.Params, param)
			}
		case paramLinkRel:
			if len(param.Values) == 1 {
				link.Rel = param.Values[0]
			} else {
				link.Params = append(link.Params, param)
			}
		case paramFmtType:
			if len(param.Values) == 1 {
				link.FmtType = param.Values[0]
			} else {
				link.Params = append(link.Params, param)
			}
		case paramLabel:
			if len(param.Values) == 1 {
				link.Label = param.Values[0]
			} else {
				link.Params = append(link.Params, param)
			}
		case paramLanguage:
			if len(param.Values) == 1 {
				link.Language = param.Values[0]
			} else {
				link.Params = append(link.Params, param)
			}
		default:
			link.Params = append(link.Params, param)
		}
	}
	return link
}
