package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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

func (dbh *DBHandler) GetProductInfo(orderID string) ([]ProductInfo, error) {
	var products []ProductInfo

	rows, err := dbh.db.Query(`
		SELECT product_id, count
		FROM orders
		WHERE order_id = $1
	`, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pi ProductInfo
		pi.OrderID = orderID

		if err := rows.Scan(&pi.Id, &pi.Count); err != nil {
			return nil, err
		}

		productRow := dbh.db.QueryRow(`
			SELECT name
			FROM product
			WHERE id = $1
		`, pi.Id)
		if err != nil {
			return nil, err
		}

		if err := productRow.Scan(&pi.Name); err != nil {
			return nil, err
		}

		productShelveRow, err := dbh.db.Query(`
			SELECT shelve_id, is_main
			FROM shelve_product
			WHERE product_id = $1
		`, pi.Id)
		if err != nil {
			return nil, err
		}
		defer productShelveRow.Close()
		for productShelveRow.Next() {
			var isMain bool
			var shelveId int

			if err := productShelveRow.Scan(&shelveId, &isMain); err != nil {
				return nil, err
			}

			shelveRow := dbh.db.QueryRow(`
				SELECT name
				FROM shelve
				WHERE id = $1
			`, shelveId)
			if err != nil {
				return nil, err
			}

			var shelveName string
			if err := shelveRow.Scan(&shelveName); err != nil {
				return nil, err
			}

			if isMain {
				pi.MainShelf = shelveName
			} else {
				pi.Shelves = append(pi.Shelves, shelveName)
			}
		}

		products = append(products, pi)
	}

	return products, nil
}

func main() {
	dbh, err := NewDBHandler()
	if err != nil {
		log.Fatal(err)
	}
	defer dbh.Close()

	IDs := os.Args[1] // Получаем номера заказов из аргументов командной строки
	orderIDs := strings.Split(IDs, ",")

	shelfProducts := make(map[string][]ProductInfo)

	for _, orderID := range orderIDs {
		products, err := dbh.GetProductInfo(orderID)
		if err != nil {
			log.Fatal(err)
		}

		for _, product := range products {
			shelfProducts[product.MainShelf] = append(shelfProducts[product.MainShelf], product)
		}
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
