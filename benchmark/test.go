package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // импорт для включения pprof
	"log"
	"time"
)

func main() {
	// Запуск профилирования на стандартном порту
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil)) // Профилирование будет доступно на порту 6060
	}()

	// Ваш основной код программы
	for {
		time.Sleep(time.Second)
		fmt.Println("Программа работает")
	}
}