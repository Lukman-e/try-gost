package base

// RequestGetAll struct used for request getAll controller funcs
type RequestGetAll struct {
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
	Keyword string `query:"search"`
	Sort    string `query:"sort"`
}
