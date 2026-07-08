package generator

import "github.com/xuri/excelize/v2"

type Styles struct {
	Header        int
	Normal        int
	Weekend       int
	PublicHoliday int
	SchoolHoliday int
	TimeFormat    int
	TimeWeekend   int
	TimeHoliday   int
	TimeSchool    int

	headerDef        excelize.Style
	normalDef        excelize.Style
	weekendDef       excelize.Style
	publicHolidayDef excelize.Style
	schoolHolidayDef excelize.Style
	timeFormatDef    excelize.Style
	timeWeekendDef   excelize.Style
	timeHolidayDef   excelize.Style
	timeSchoolDef    excelize.Style
}

func createStyles(f *excelize.File) (*Styles, error) {
	s := &Styles{}
	var err error

	s.headerDef = excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "000000", Style: 1},
		},
	}
	s.Header, err = f.NewStyle(&s.headerDef)
	if err != nil {
		return nil, err
	}

	s.normalDef = excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	}
	s.Normal, err = f.NewStyle(&s.normalDef)
	if err != nil {
		return nil, err
	}

	s.weekendDef = excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	}
	s.Weekend, err = f.NewStyle(&s.weekendDef)
	if err != nil {
		return nil, err
	}

	s.publicHolidayDef = excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FCE4D6"}},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	}
	s.PublicHoliday, err = f.NewStyle(&s.publicHolidayDef)
	if err != nil {
		return nil, err
	}

	s.schoolHolidayDef = excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	}
	s.SchoolHoliday, err = f.NewStyle(&s.schoolHolidayDef)
	if err != nil {
		return nil, err
	}

	s.timeFormatDef = excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	}
	s.TimeFormat, err = f.NewStyle(&s.timeFormatDef)
	if err != nil {
		return nil, err
	}

	s.timeWeekendDef = excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	}
	s.TimeWeekend, err = f.NewStyle(&s.timeWeekendDef)
	if err != nil {
		return nil, err
	}

	s.timeHolidayDef = excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FCE4D6"}},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	}
	s.TimeHoliday, err = f.NewStyle(&s.timeHolidayDef)
	if err != nil {
		return nil, err
	}

	s.timeSchoolDef = excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	}
	s.TimeSchool, err = f.NewStyle(&s.timeSchoolDef)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Styles) withBorder(f *excelize.File, baseStyle int, top, bottom, left, right bool) int {
	var base excelize.Style
	switch baseStyle {
	case s.Header:
		base = s.headerDef
	case s.Normal:
		base = s.normalDef
	case s.Weekend:
		base = s.weekendDef
	case s.PublicHoliday:
		base = s.publicHolidayDef
	case s.SchoolHoliday:
		base = s.schoolHolidayDef
	case s.TimeFormat:
		base = s.timeFormatDef
	case s.TimeWeekend:
		base = s.timeWeekendDef
	case s.TimeHoliday:
		base = s.timeHolidayDef
	case s.TimeSchool:
		base = s.timeSchoolDef
	default:
		return baseStyle
	}

	var borders []excelize.Border
	borders = append(borders, base.Border...)
	if top {
		borders = append(borders, excelize.Border{Type: "top", Color: "000000", Style: 1})
	}
	if bottom {
		borders = append(borders, excelize.Border{Type: "bottom", Color: "000000", Style: 1})
	}
	if left {
		borders = append(borders, excelize.Border{Type: "left", Color: "000000", Style: 1})
	}
	if right {
		borders = append(borders, excelize.Border{Type: "right", Color: "000000", Style: 1})
	}
	base.Border = borders

	id, _ := f.NewStyle(&base)
	return id
}

func strPtr(s string) *string {
	return &s
}
