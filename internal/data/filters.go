package data

import (
	"strings"

	"greenlight.ilx.net/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

func (f Filters) sortCoulmn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirecetion() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page <= 10_000_000 && f.Page >= 1, "page", "page must be between 1 and 10,000,000")
	v.Check(f.PageSize <= 100 && f.Page >= 1, "page_size", "page size must be between 1 and 100")
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

func (f Filters) Limit() int {
	return f.PageSize
}

func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}

func CalculateMetadata(totalrecords, page, pageSize int) Metadata {
	if totalrecords == 0 {
		return Metadata{}
	}

	return Metadata{
		PageSize:     pageSize,
		CurrentPage:  page,
		FirstPage:    1,
		LastPage:     (totalrecords + pageSize - 1) / pageSize,
		TotalRecords: totalrecords,
	}

}
