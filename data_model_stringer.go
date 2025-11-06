package iec61850

import "strings"

// Pretty Stringers with indentation starting at top node "DataModel"

// indent returns a string with two-space indentation repeated level times.
func indent(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat("  ", level)
}

// DataModel stringer prints the full hierarchical data model.
func (dm DataModel) String() string {
	var b strings.Builder
	b.WriteString("DataModel\n")
	for _, ld := range dm.LDs {
		ld.writeTo(&b, 1)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (ld LD) String() string {
	var b strings.Builder
	ld.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (ld LD) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "LD: " + ld.Data + "\n")
	for _, ln := range ld.LNs {
		ln.writeTo(b, level+1)
	}
}

func (ln LN) String() string {
	var b strings.Builder
	ln.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (ln LN) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "LN: " + ln.Data + "\n")
	for _, d := range ln.DOs {
		d.writeTo(b, level+1)
	}
	for _, ds := range ln.DSs {
		ds.writeTo(b, level+1)
	}
	for _, r := range ln.URReports {
		r.writeTo(b, level+1)
	}
	for _, r := range ln.BRReports {
		r.writeTo(b, level+1)
	}
}

func (r URReport) String() string {
	var b strings.Builder
	r.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (r URReport) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "URReport: " + r.Data + "\n")
}

func (r BRReport) String() string {
	var b strings.Builder
	r.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (r BRReport) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "BRReport: " + r.Data + "\n")
}

func (ds DS) String() string {
	var b strings.Builder
	ds.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (ds DS) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "DS: " + ds.Data)
	if ds.IsDeletable {
		b.WriteString(" (deletable)\n")
	} else {
		b.WriteString(" (not deletable)\n")
	}
	for _, ref := range ds.DSRefs {
		ref.writeTo(b, level+1)
	}
}

func (ref DSRef) String() string {
	var b strings.Builder
	ref.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (ref DSRef) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "DSRef: " + ref.Data + "\n")
}

func (d DO) String() string {
	var b strings.Builder
	d.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (d DO) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "DO: " + d.Data + "\n")
	for _, da := range d.DAs {
		da.writeTo(b, level+1)
	}
}

func (da DA) String() string {
	var b strings.Builder
	da.writeTo(&b, 0)
	return strings.TrimRight(b.String(), "\n")
}

func (da DA) writeTo(b *strings.Builder, level int) {
	b.WriteString(indent(level) + "DA: " + da.Data + "\n")
	for _, child := range da.DAs {
		child.writeTo(b, level+1)
	}
}
