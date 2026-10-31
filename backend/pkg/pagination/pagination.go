package pagination

import "strconv"

// Params holds pagination parameters from HTTP requests.
type Params struct {
	Page    int
	PerPage int
}

// Parse parses pagination parameters from a query string.
func Parse(pageStr, perPageStr string) Params {
	page := 1
	perPage := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr != "" {
		if p, err := strconv.Atoi(perPageStr); err == nil && p > 0 {
			perPage = p
		}
	}

	// Enforce max page size
	if perPage > 100 {
		perPage = 100
	}

	return Params{Page: page, PerPage: perPage}
}

// Offset returns the SQL offset for the given page and perPage.
func Offset(page, perPage int) int {
	return (page - 1) * perPage
}

// TotalPages calculates total pages from total items and perPage.
func TotalPages(total, perPage int64) int {
	pages := total / int64(perPage)
	if total%int64(perPage) > 0 {
		pages++
	}
	return int(pages)
}
