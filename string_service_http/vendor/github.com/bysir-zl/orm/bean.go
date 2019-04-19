package orm


type Page struct {
	Total     int64 `json:"total,omitempty"`
	PageTotal int   `json:"page_total,omitempty"`
	Page      int   `json:"page,omitempty"`
	PageSize  int   `json:"page_size,omitempty"`
}
