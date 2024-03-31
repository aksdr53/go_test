package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

type DBHandler struct {
	db *sql.DB
}

func NewDBHandler() (*DBHandler, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	return &DBHandler{db: db}, nil
}

func (dbh *DBHandler) Close() {
	dbh.db.Close()
}

type ProductInfo struct {
	Name      string
	Id        int
	OrderID   string
	Count     int
	Shelves   []string
	MainShelf string
}

type Shelves struct {
	shelveId   int
	shelveName string
	productId  int
	isMain     bool
}

func (dbh *DBHandler) GetProductInfo(orderIDs string) ([]ProductInfo, error) {

	var pis []ProductInfo

	Names := make(map[int]string)
	var stringIds []string

	var shelves []Shelves
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
		var pi ProductInfo

		if err := rows.Scan(&pi.Id, &pi.Count, &pi.OrderID); err != nil {
			return nil, err
		}

		stringIds = append(stringIds, strconv.Itoa(pi.Id)) //добавить избегание дублирования
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

		if err := productShelveRow.Scan(&sh.shelveId, &sh.isMain, &sh.productId); err != nil {
			return nil, err
		}
		stringShelves = append(stringShelves, strconv.Itoa(sh.shelveId))
		shelves = append(shelves, sh)
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

		for i, shelve := range shelves {
			if shelve.shelveId == shelveId {

				shelves[i].shelveName = shelveName

			}

		}
	}

	for i, pi := range pis {
		for _, shelve := range shelves {
			if pi.Id == shelve.productId {

				if shelve.isMain {
					pis[i].MainShelf = shelve.shelveName

				} else {
					pis[i].Shelves = append(pis[i].Shelves, shelve.shelveName)
				}
			}
		}
		pis[i].Name = Names[pi.Id]

	}

	return pis, nil
}

func main() {
	dbh, err := NewDBHandler()
	if err != nil {
		log.Fatal(err)
	}
	defer dbh.Close()

	orderIDs := os.Args[1] // Получаем номера заказов из аргументов командной строки

	shelfProducts := make(map[string][]ProductInfo)

	products, err := dbh.GetProductInfo(orderIDs)

	if err != nil {
		log.Fatal(err)
	}

	for _, product := range products {
		shelfProducts[product.MainShelf] = append(shelfProducts[product.MainShelf], product)
	}

	for shelf, products := range shelfProducts {
		fmt.Printf("===Стеллаж %s\n", shelf)
		for _, product := range products {
			fmt.Printf("%s (id=%d)\nзаказ %s, %d шт\n", product.Name, product.Id, product.OrderID, product.Count)
			if len(product.Shelves) > 0 {
				fmt.Printf("доп стеллаж: %s\n", strings.Join(product.Shelves, ","))
			}
		}
	}
}
