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
}

func createStyles(f *excelize.File) (*Styles, error) {
	s := &Styles{}
	var err error

	s.Header, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	s.Normal, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return nil, err
	}

	s.Weekend, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return nil, err
	}

	s.PublicHoliday, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FCE4D6"}},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return nil, err
	}

	s.SchoolHoliday, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return nil, err
	}

	s.TimeFormat, err = f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	})
	if err != nil {
		return nil, err
	}

	s.TimeWeekend, err = f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9D9D9"}},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	})
	if err != nil {
		return nil, err
	}

	s.TimeHoliday, err = f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FCE4D6"}},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	})
	if err != nil {
		return nil, err
	}

	s.TimeSchool, err = f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 10},
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}},
		Alignment:    &excelize.Alignment{Vertical: "center"},
		CustomNumFmt: strPtr("[h]:mm"),
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func strPtr(s string) *string {
	return &s
}
