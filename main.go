package main

import (
	"encoding/json"               // Пакет для работы с JSON-данными
	"github.com/labstack/echo/v4" // Веб-фреймворк Echo
	"net"                         // Пакет для работы с сетью
	"net/http"                    // Пакет для работы с HTTP-протоколом
	"os"                          // Пакет для работы с операционной системой
	"sync"                        // Пакет для работы с синхронизацией
	"time"                        // Пакет для работы с временем
)

type Server struct {
	IP string `json:"ip"` // Структура для хранения IP-адреса сервера
}

type PingResult struct {
	IP     string `json:"ip"`     // Структура для хранения IP-адреса сервера в результате пинга
	Status string `json:"status"` // Структура для хранения статуса пинга (Ok или Error)
}

type PingResults struct {
	sync.Mutex                   // Структура для синхронизации доступа к результатам пинга
	Results    map[string]string `json:"results"` // Структура для хранения результатов пинга (IP-адрес -> статус)
}

func main() {
	// Чтение JSON-файла со списком серверов
	serversFile, err := ReadFile("servers.json") // Чтение JSON-файла со списком серверов
	if err != nil {
		panic(err) // В случае ошибки - завершение программы с ошибкой
	}

	// Декодирование JSON-файла
	var servers []Server                                          // Объявление переменной для хранения списка серверов
	if err := json.Unmarshal(serversFile, &servers); err != nil { // Декодирование JSON-файла и сохранение результатов в переменной servers
		panic(err) // В случае ошибки - завершение программы с ошибкой
	}

	// Создание экземпляра Echo
	e := echo.New() // Создание экземпляра веб-фреймворка Echo

	// Обработчик для пингования всех серверов
	e.GET("/ping", func(c echo.Context) error {
		// Инициализация переменных
		var wg sync.WaitGroup    // Инициализация объекта для ожидания выполнения группы горутин
		results := &PingResults{ // Инициализация объекта для хранения результатов пинга
			Results: make(map[string]string),
		}

		// Параллельное пингование всех серверов
		for _, s := range servers { // Итерация по списку серверов
			wg.Add(1)            // Добавление горутины в группу
			go func(ip string) { // Запуск горутины для пингования сервера с указанным IP-адресом
				status := "Ok" // Инициализация статуса пинга
				conn, err := net.DialTimeout("tcp", ip+":80", 2*time.Second)
				if err != nil {
					// Если соединение не удалось, статус устанавливается в "Error"
					status = "Error"
				} else {
					// Если соединение удалось, соединение закрывается
					conn.Close()
				}

				// Блокировка для доступа к общим результатам
				results.Lock()
				// Сохранение статуса в общем результате
				results.Results[ip] = status
				// Разблокировка доступа к общим результатам
				results.Unlock()
				// Сигнал о завершении работы горутины
				wg.Done()
			}(s.IP)
		}

		// Ожидание завершения всех пингов
		wg.Wait()

		// Возврат результата в формате JSON
		return c.JSON(http.StatusOK, results)
	})

	// Запуск сервера
	e.Logger.Fatal(e.Start(":8080"))
}

// ReadFile читает файл по указанному пути и возвращает его содержимое в виде []byte

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path) // Открываем файл по указанному пути
	if err != nil {
		return nil, err // Возвращаем ошибку, если произошла
	}
	defer file.Close() // Отложенное закрытие файла

	stat, err := file.Stat() // Получаем информацию о файле
	if err != nil {
		return nil, err // Возвращаем ошибку, если произошла
	}

	buf := make([]byte, stat.Size()) // Создаем буфер с размером файла
	_, err = file.Read(buf)          // Читаем содержимое файла в буфер
	if err != nil {
		return nil, err // Возвращаем ошибку, если произошла
	}

	return buf, nil // Возвращаем содержимое файла в виде []byte и ошибку - nil, если таковая отсутствует
}
