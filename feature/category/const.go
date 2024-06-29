package category

var Categories = map[uint8]Category{
	1: {ID: 1, Name: "VIP", Price: 3_800_000},
	2: {ID: 2, Name: "PLATINUM", Price: 3_400_000},
	3: {ID: 3, Name: "CAT 1", Price: 2_900_000},
	4: {ID: 4, Name: "CAT 2", Price: 2_600_000},
	5: {ID: 5, Name: "CAT 3", Price: 2_100_000},
}

const (
	listCategoriesCacheKey = "categories"
)
