package filters

import (
	"errors"
	"strings"
)

const (
	AscSort  = "ASC"
	DescSort = "DESC"
)

type Filters struct {
	Page int 
	PageSize int
	Sort string
	SortSafelist []string
}

func (f *Filters) SortColumn() string {
	s := strings.TrimPrefix(f.Sort, "-")
	for _, safeValue := range f.SortSafelist {
		if strings.EqualFold(s, safeValue) {
			return s
		}
	}
	panic(errors.New("Unknown sort column: " + f.Sort))
}

func (f *Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return DescSort
	}
	return AscSort
}