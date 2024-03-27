package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Shelve struct {
	ID   int
	Name string
}

type Product struct {
	ID         int
	Name       string
	MainShelve int
}

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orderIDs := os.Args[1:] // Получаем номера заказов из аргументов командной строки

	for _, orderID := range orderIDs {
		products := getProductsInOrder(db, orderID)

		for _, product := range products {
			shelves := getShelvesForProduct(db, product.ID)
			fmt.Printf("Для товара %s из заказа %s необходимо посетить следующие стеллажи:\n", product.Name, orderID)
			for _, sh := range shelves {
				fmt.Printf("%s\n", sh.Name)
			}
		}
	}
}

func getProductsInOrder(db *sql.DB, orderID string) []Product {
	var products []Product

	rows, err := db.Query("SELECT id, name, main_shelve FROM Товар WHERE id IN (SELECT id_товара FROM Заказы WHERE id_заказа = $1)", orderID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.MainShelve); err != nil {
			log.Fatal(err)
		}
		products = append(products, p)
	}

	return products
}

func getShelvesForProduct(db *sql.DB, productID int) []Shelve {
	var shelves []Shelve

	rows, err := db.Query("SELECT s.id, s.name FROM Стеллаж s JOIN Прочие_стеллажи ps ON s.id = ps.id_стеллажа WHERE ps.id_товара = $1", productID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s Shelve
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			log.Fatal(err)
		}
		shelves = append(shelves, s)
	}

	return shelves
}


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
	user     = "youruser"
	password = "yourpassword"
	dbname   = "yourdbname"
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

	fmt.Println("Введите номера заказов через запятую (например, 10,11,14):")
	var input string
	fmt.Scanln(&input)

	orderIDs := strings.Split(input, ",")

	type ProductInfo struct {
		Name        string
		OrderID     string
		Count       int
		Shelves     []string
		MainShelf   string
		Additional  string
	}

	shelfProducts := make(map[string][]ProductInfo)

	for _, orderID := range orderIDs {
		rows, err := db.Query(`
			SELECT p.name, o.id, o.count, s.name, p.main_shelve
			FROM orders o
			JOIN products p ON o.product_id = p.id
			JOIN shelves s ON p.main_shelve = s.id
			WHERE o.id = $1
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
				FROM additional_shelves a
				JOIN shelves s ON a.shelve_id = s.id
				WHERE a.product_id = (SELECT id FROM products WHERE name = $1)
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