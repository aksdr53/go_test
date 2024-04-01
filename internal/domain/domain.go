package domain

type ProductInfo struct {
	Name      string
	Id        int
	OrderID   string
	Count     int
	Shelves   []string
	MainShelf string
}

type Shelves struct {
	shelveId int
	isMain   bool
}

type ProductRepository interface {
	GetProductInfo(orderIDs string) ([]ProductInfo, error)
}
