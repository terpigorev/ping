package main

import (
	"encoding/json"               // Пакет для кодирования и декодирования JSON
	"fmt"                         // Пакет для форматированного ввода-вывода
	"github.com/labstack/echo/v4" // Внешний пакет для создания веб-приложений
	"net"                         // Пакет для операций сети
	"net/http"                    // Пакет для функциональности HTTP-клиента и HTTP-сервера
	"os"                          // Пакет предоставляет возможность использования функций, зависящих от операционной системы
	"sync"                        // Пакет для предоставления базовых примитивов синхронизации, таких как мьютексы
	"time"                        // Пакет для работы со временем
)

type Server struct {
	IP string `json:"ip"` // Структура, представляющая сервер с IP-адресом
}

type PingResult struct {
	IP     string `json:"ip"`     // Структура, представляющая IP-адрес пингуемого сервера
	Status string `json:"status"` // Структура, представляющая статус пинга (например, Ok или Error)
}

type PingResults struct {
	sync.Mutex
	Results map[string]string `json:"results"` // Структура, представляющая результаты пингования нескольких серверов
}

func main() {
	pingServers := func(servers []Server) *PingResults { // Функция для пингования нескольких серверов и сбора результатов
		var wg sync.WaitGroup
		results := &PingResults{
			Results: make(map[string]string),
		}

		for _, s := range servers {
			wg.Add(1)
			go func(ip string) {
				status := "Ok"
				conn, err := net.DialTimeout("tcp", ip+":80", 2*time.Second) // Установка TCP-соединения с сервером
				if err != nil {
					status = "Error"
				} else {
					conn.Close()
				}

				results.Lock()
				results.Results[ip] = status // Сохранение статуса пинга сервера
				results.Unlock()
				wg.Done()
			}(s.IP)
		}

		wg.Wait()
		return results
	}

	serversFile, err := ReadFile("servers.json") // Чтение файла конфигурации серверов
	if err != nil {
		panic(err)
	}

	var servers []Server
	if err := json.Unmarshal(serversFile, &servers); err != nil { // Распаковка JSON-данных в структуру серверов
		panic(err)
	}

	e := echo.New() // Создание нового экземпляра веб-фреймворка Echo

	e.GET("/ping", func(c echo.Context) error { // Обработка GET-запроса к "/ping"
		results := pingServers(servers) // Пингование серверов и получение результатов

		fmt.Println("Результаты пинга:")
		for ip, status := range results.Results { // Вывод результатов пинга
			fmt.Printf("Сервер: %s, Статус: %s\n", ip, status)
		}

		go func() { // Запуск горутины для периодического пингования серверов
			for range time.Tick(10 * time.Second) {
				results := pingServers(servers)
				fmt.Println("Результаты пинга:")
				for ip, status := range results.Results {
					fmt.Printf("Сервер: %s, Статус: %s\n", ip, status)
				}
			}
		}()

		return c.JSON(http.StatusOK, results) // Возврат результатов пинга в виде JSON-ответа
	})

	e.Logger.Fatal(e.Start(":8080")) // Запуск сервера и прослушивание порта 8080
}

func ReadFile(path string) ([]byte, error) { // Функция для чтения содержимого файла
	file, err := os.Open(path) // Открытие файла
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat() // Получение информации о файле
	if err != nil {
		return nil, err
	}

	buf := make([]byte, stat.Size())
	_, err = file.Read(buf) // Чтение данных из файла
	if err != nil {
		return nil, err
	}

	return buf, nil
}
