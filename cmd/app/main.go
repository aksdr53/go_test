package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"go_test/internal/domain"
	"go_test/internal/repository"
	"go_test/internal/usecase"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", constructDSN())
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	dbHandler := repository.NewDBHandler(db)
	productUseCase := usecase.NewProductUseCase(dbHandler)

	orderIDs := os.Args[1] // Получаем номера заказов из аргументов командной строки
	products, err := productUseCase.FetchProductInfo(orderIDs)
	if err != nil {
		log.Fatal(err)
	}
	shelfProducts := make(map[string][]domain.ProductInfo)
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

func constructDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return psqlInfo
}
