package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "go_user"
	password = "password"
	dbname   = "go_db"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	IDs := os.Args[1] // Получаем номера заказов из аргументов командной строки

	orderIDs := strings.Split(IDs, ",")

	type ProductInfo struct {
		Name       string
		OrderID    string
		Count      int
		Shelves    []string
		MainShelf  string
		Additional string
	}

	shelfProducts := make(map[string][]ProductInfo)

	for _, orderID := range orderIDs {
		rows, err := db.Query(`
			SELECT p.name, o.order_id, o.count, s.name, p.main_shelve
			FROM orders o
			JOIN product p ON o.product_id = p.id
			JOIN shelve s ON p.main_shelve = s.id
			WHERE o.order_id = $1
		`, strings.TrimSpace(orderID))
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var pi ProductInfo
			var shelfName string
			if err := rows.Scan(&pi.Name, &pi.OrderID, &pi.Count, &shelfName, &pi.MainShelf); err != nil {
				log.Fatal(err)
			}

			// Get additional shelves
			additionalRows, err := db.Query(`
				SELECT s.name
				FROM additional_shelve a
				JOIN shelve s ON a.shelve_id = s.id
				WHERE a.product_id = (SELECT id FROM product WHERE name = $1)
			`, pi.Name)
			if err != nil {
				log.Fatal(err)
			}
			defer additionalRows.Close()

			for additionalRows.Next() {
				var additionalShelf string
				if err := additionalRows.Scan(&additionalShelf); err != nil {
					log.Fatal(err)
				}
				pi.Shelves = append(pi.Shelves, additionalShelf)
			}

			shelfProducts[shelfName] = append(shelfProducts[shelfName], pi)
		}
	}

	for shelf, products := range shelfProducts {
		fmt.Printf("===Стеллаж %s\n", shelf)
		for _, product := range products {
			fmt.Printf("%s (id=%s)\nзаказ %s, %d шт\n", product.Name, product.MainShelf, product.OrderID, product.Count)
			if len(product.Shelves) > 0 {
				fmt.Printf("доп стеллаж: %s\n", strings.Join(product.Shelves, ","))
			}
		}
	}
}
