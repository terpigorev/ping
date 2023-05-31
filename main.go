package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Server struct {
	IP string `json:"ip"`
}

type PingResult struct {
	IP     string `json:"ip"`
	Status string `json:"status"`
}

type PingResults struct {
	sync.Mutex
	Results map[string]string `json:"results"`
}

func main() {
	pingServers := func(servers []Server) *PingResults {
		var wg sync.WaitGroup
		results := &PingResults{
			Results: make(map[string]string),
		}

		for _, s := range servers {
			wg.Add(1)
			go func(ip string) {
				status := "Ok"
				conn, err := net.DialTimeout("tcp", ip+":80", 2*time.Second)
				if err != nil {
					status = "Error"
				} else {
					conn.Close()
				}

				results.Lock()
				results.Results[ip] = status
				results.Unlock()
				wg.Done()
			}(s.IP)
		}

		wg.Wait()
		return results
	}

	serversFile, err := ReadFile("servers.json")
	if err != nil {
		panic(err)
	}

	var servers []Server
	if err := json.Unmarshal(serversFile, &servers); err != nil {
		panic(err)
	}

	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		results := pingServers(servers)

		fmt.Println("Ping Results:")
		for ip, status := range results.Results {
			fmt.Printf("Server: %s, Status: %s\n", ip, status)
		}

		go func() {
			for range time.Tick(10 * time.Second) {
				results := pingServers(servers)
				fmt.Println("Ping Results:")
				for ip, status := range results.Results {
					fmt.Printf("Server: %s, Status: %s\n", ip, status)
				}
			}
		}()

		return c.JSON(http.StatusOK, results)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, stat.Size())
	_, err = file.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
