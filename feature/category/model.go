package category

type listCategoriesResponse struct {
	ID    uint8  `json:"id"`
	Name  string `json:"name,omitempty"`
	Price uint32 `json:"price,omitempty"`
	Total uint16 `json:"total"`
}
