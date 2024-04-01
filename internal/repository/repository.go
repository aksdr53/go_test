package repository

import (
	"database/sql"
	"strconv"
	"strings"

	"go_test/internal/domain"
)

type DBHandler struct {
	db *sql.DB
}

func NewDBHandler(db *sql.DB) *DBHandler {
	return &DBHandler{db: db}
}

type Shelves struct {
	shelveId int
	isMain   bool
}

func (dbh *DBHandler) GetProductInfo(orderIDs string) ([]domain.ProductInfo, error) {
	var pis []domain.ProductInfo
	shs := make(map[int][]Shelves)
	shelveNames := make(map[int]string)
	Names := make(map[int]string)
	var stringIds []string
	var stringShelves []string

	rows, err := dbh.db.Query(`
		SELECT product_id, count, order_id
		FROM orders
		WHERE order_id IN (
	` + orderIDs + `)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pi domain.ProductInfo

		if err := rows.Scan(&pi.Id, &pi.Count, &pi.OrderID); err != nil {
			return nil, err
		}

		stringIds = append(stringIds, strconv.Itoa(pi.Id))
		pis = append(pis, pi)
	}
	idsString := strings.Join(stringIds, ",")

	productRows, err := dbh.db.Query(`
		SELECT name, id
		FROM product
		WHERE id IN (
	` + idsString + `)`)
	if err != nil {
		return nil, err
	}

	for productRows.Next() {
		var id int
		var Name string
		if err := productRows.Scan(&Name, &id); err != nil {
			return nil, err
		}

		Names[id] = Name
	}

	productShelveRow, err := dbh.db.Query(`
		SELECT shelve_id, is_main, product_id
		FROM shelve_product
		WHERE product_id IN (
	` + idsString + `)`)
	if err != nil {
		return nil, err
	}
	defer productShelveRow.Close()
	for productShelveRow.Next() {
		var sh Shelves
		var productId int

		if err := productShelveRow.Scan(&sh.shelveId, &sh.isMain, &productId); err != nil {
			return nil, err
		}
		stringShelves = append(stringShelves, strconv.Itoa(sh.shelveId))
		shs[productId] = append(shs[productId], sh)
	}
	shelveIds := strings.Join(stringShelves, ",")

	shelveRows, err := dbh.db.Query(`
		SELECT name, id
		FROM shelve
		WHERE id IN (
	` + shelveIds + `)`)
	if err != nil {
		return nil, err
	}
	defer shelveRows.Close()
	for shelveRows.Next() {
		var shelveName string
		var shelveId int
		if err := shelveRows.Scan(&shelveName, &shelveId); err != nil {
			return nil, err
		}
		shelveNames[shelveId] = shelveName

	}

	for i, pi := range pis {
		pis[i].Name = Names[pi.Id]
		for _, sh := range shs[pi.Id] {
			shelveId := sh.shelveId
			if sh.isMain {
				pis[i].MainShelf = shelveNames[shelveId]
			} else {
				pis[i].Shelves = append(pis[i].Shelves, shelveNames[shelveId])
			}
		}

	}

	return pis, nil
}
