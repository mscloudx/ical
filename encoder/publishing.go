package encoder

import (
	"strings"

	"github.com/mscloudx/ical/model"
)

// writeParticipant записывает PARTICIPANT.
func (w *icsWriter) writeParticipant(p *model.Participant) {
	w.writeBegin("PARTICIPANT")

	w.writePropStr("UID", p.UID)
	w.writeDTStamp(p.DTStamp)

	if p.ParticipantType != "" {
		w.writePropStr("PARTICIPANT-TYPE", p.ParticipantType)
	}
	if p.CalendarAddress != "" {
		w.writePropStr("CALENDAR-ADDRESS", p.CalendarAddress)
	}
	if p.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*p.Created, false))
	}
	w.writePropStr("DESCRIPTION", p.Description)

	if p.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*p.LastModified, false))
	}
	if p.Sequence != 0 {
		w.writePropInt("SEQUENCE", p.Sequence)
	}
	w.writePropStr("SUMMARY", p.Summary)
	w.writePropStr("URL", p.URL)

	if len(p.Categories) > 0 {
		w.writePropStr("CATEGORIES", strings.Join(p.Categories, ","))
	}
	for i := range p.Contacts {
		w.writePropStr("CONTACT", p.Contacts[i])
	}
	if p.Location != "" {
		w.writePropStr("LOCATION", p.Location)
	}

	for i := range p.Concepts {
		w.writePropStr("CONCEPT", p.Concepts[i])
	}
	for i := range p.RefIDs {
		w.writePropStr("REFID", p.RefIDs[i])
	}
	for i := range p.Links {
		w.writeLink(&p.Links[i])
	}

	w.writeProperties(p.XProps)
	w.writeProperties(p.IanaProps)

	for i := range p.StructuredData {
		w.writeStructuredData(&p.StructuredData[i])
	}

	for i := range p.Locations {
		w.writeLocationComponent(&p.Locations[i])
	}
	for i := range p.Resources {
		w.writeResourceComponent(&p.Resources[i])
	}

	w.writeEnd("PARTICIPANT")
}

// writeLocationComponent записывает LOCATION.
func (w *icsWriter) writeLocationComponent(l *model.LocationComponent) {
	w.writeBegin("LOCATION")

	w.writePropStr("UID", l.UID)
	w.writeDTStamp(l.DTStamp)
	w.writePropStr("NAME", l.Name)
	w.writePropStr("DESCRIPTION", l.Description)
	w.writePropStr("GEO", l.Geo)

	if len(l.Type) > 0 {
		w.writePropStr("LOCATION-TYPE", strings.Join(l.Type, ","))
	}

	if l.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*l.Created, false))
	}
	if l.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*l.LastModified, false))
	}
	if l.Sequence != 0 {
		w.writePropInt("SEQUENCE", l.Sequence)
	}
	w.writePropStr("URL", l.URL)

	for i := range l.Concepts {
		w.writePropStr("CONCEPT", l.Concepts[i])
	}
	for i := range l.RefIDs {
		w.writePropStr("REFID", l.RefIDs[i])
	}
	for i := range l.Links {
		w.writeLink(&l.Links[i])
	}

	w.writeProperties(l.XProps)
	w.writeProperties(l.IanaProps)

	for i := range l.StructuredData {
		w.writeStructuredData(&l.StructuredData[i])
	}

	w.writeEnd("LOCATION")
}

// writeResourceComponent записывает RESOURCE.
func (w *icsWriter) writeResourceComponent(r *model.ResourceComponent) {
	w.writeBegin("RESOURCE")

	w.writePropStr("UID", r.UID)
	w.writeDTStamp(r.DTStamp)
	w.writePropStr("NAME", r.Name)
	w.writePropStr("DESCRIPTION", r.Description)

	if len(r.ResourceType) > 0 {
		w.writePropStr("RESOURCE-TYPE", strings.Join(r.ResourceType, ","))
	}

	if r.Created != nil {
		w.writePropStr("CREATED", formatDateTime(*r.Created, false))
	}
	if r.LastModified != nil {
		w.writePropStr("LAST-MODIFIED", formatDateTime(*r.LastModified, false))
	}
	if r.Sequence != 0 {
		w.writePropInt("SEQUENCE", r.Sequence)
	}
	w.writePropStr("URL", r.URL)

	for i := range r.Concepts {
		w.writePropStr("CONCEPT", r.Concepts[i])
	}
	for i := range r.RefIDs {
		w.writePropStr("REFID", r.RefIDs[i])
	}
	for i := range r.Links {
		w.writeLink(&r.Links[i])
	}

	w.writeProperties(r.XProps)
	w.writeProperties(r.IanaProps)

	for i := range r.StructuredData {
		w.writeStructuredData(&r.StructuredData[i])
	}

	w.writeEnd("RESOURCE")
}

// writeStructuredData записывает STRUCTURED-DATA.
func (w *icsWriter) writeStructuredData(sd *model.StructuredData) {
	prop := model.Property{
		Name:  "STRUCTURED-DATA",
		Value: sd.Value,
	}
	if sd.ValueType != "" {
		prop.Params = append(prop.Params, model.Param{Name: "VALUE", Values: []string{sd.ValueType}})
	}
	if sd.FmtType != "" {
		prop.Params = append(prop.Params, model.Param{Name: "FMTTYPE", Values: []string{sd.FmtType}})
	}
	if sd.Schema != "" {
		prop.Params = append(prop.Params, model.Param{Name: "SCHEMA", Values: []string{sd.Schema}})
	}
	w.writeProperty(prop)
}

// writeStyledDescription записывает STYLED-DESCRIPTION.
func (w *icsWriter) writeStyledDescription(sd *model.StyledDescription) {
	prop := model.Property{
		Name:  "STYLED-DESCRIPTION",
		Value: sd.Value,
	}
	if sd.ValueType != "" {
		prop.Params = append(prop.Params, model.Param{Name: "VALUE", Values: []string{sd.ValueType}})
	}
	if sd.FmtType != "" {
		prop.Params = append(prop.Params, model.Param{Name: "FMTTYPE", Values: []string{sd.FmtType}})
	}
	if sd.Language != "" {
		prop.Params = append(prop.Params, model.Param{Name: "LANGUAGE", Values: []string{sd.Language}})
	}
	w.writeProperty(prop)
}

// writeLink записывает LINK.
func (w *icsWriter) writeLink(l *model.Link) {
	prop := model.Property{
		Name:  "LINK",
		Value: l.Value,
	}
	if l.ValueType != "" {
		prop.Params = append(prop.Params, model.Param{Name: "VALUE", Values: []string{l.ValueType}})
	}
	if l.Rel != "" {
		prop.Params = append(prop.Params, model.Param{Name: "LINKREL", Values: []string{l.Rel}})
	}
	if l.FmtType != "" {
		prop.Params = append(prop.Params, model.Param{Name: "FMTTYPE", Values: []string{l.FmtType}})
	}
	if l.Label != "" {
		prop.Params = append(prop.Params, model.Param{Name: "LABEL", Values: []string{l.Label}})
	}
	if l.Language != "" {
		prop.Params = append(prop.Params, model.Param{Name: "LANGUAGE", Values: []string{l.Language}})
	}
	if len(l.Params) > 0 {
		prop.Params = append(prop.Params, l.Params...)
	}
	w.writeProperty(prop)
}
