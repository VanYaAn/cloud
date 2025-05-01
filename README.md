## Тестовое задание
## Настройка бэкендов
я использовал переменные окружения в качестве хранения данных для бэкенд серверов
- экспорт
```zsh
export NumberOfServers=6
export BackendAddres0=8080
export BackendAddres1=8081
export BackendAddres2=8082
export BackendAddres3=8083
export BackendAddres4=8084
export BackendAddres5=8085
```
## Docker-compose file 
```docker-compose.yaml
version: '3.8'

services:
  pg-local:
    image: postgres:latest
    environment:
      POSTGRES_USER: your_user
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: your_database
    ports:
      - "5432:5432"
    volumes:
      - pg-data:/var/lib/postgresql/data

volumes:
  pg-data:
```
-сборка
```zsh
docker-compose up --build -d
```
-миграция
```zsh
mkdir db/migrations
goose -dir db/migrations postgres "user=your_user password=your_password dbname=your_database host=localhost port=5432 sslmode=disable" up
```
##Тест 
```zsh
% ab -n 5000 -c 100 http://localhost:8080/
This is ApacheBench, Version 2.3 <$Revision: 1913912 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 500 requests
Completed 1000 requests
Completed 1500 requests
Completed 2000 requests
Completed 2500 requests
Completed 3000 requests
Completed 3500 requests
Completed 4000 requests
Completed 4500 requests
Completed 5000 requests
Finished 5000 requests


Server Software:        
Server Hostname:        localhost
Server Port:            8080

Document Path:          /
Document Length:        22 bytes

Concurrency Level:      100
Time taken for tests:   0.795 seconds
Complete requests:      5000
Failed requests:        0
Non-2xx responses:      5000
Total transferred:      945000 bytes
HTML transferred:       110000 bytes
Requests per second:    6289.54 [#/sec] (mean)
Time per request:       15.899 [ms] (mean)
Time per request:       0.159 [ms] (mean, across all concurrent requests)
Transfer rate:          1160.86 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    2   1.9      1      45
Processing:     0   14  21.6      7     145
Waiting:        0   13  21.5      6     145
Total:          0   16  21.6      9     146

Percentage of the requests served within a certain time (ms)
  50%      9
  66%     12
  75%     16
  80%     19
  90%     30
  95%     52
  98%    113
  99%    123
 100%    146 (longest request)
```
## Использовал стандартный логер из "log"
## Для проверки кода использова golangci-lint
## HealthCheck у меня StateCheck()
```go 
func StateCheck() {
	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			server.StateCheck()
			log.Println("Health check completed")
		}
	}
}
```
Если не работает корректно lbrl(load balancer with race limiter), то предлагаю использоватьа lb как хендлер (больше его тестировал )
## цели над улучшениями
-сдеать слои и лучше сделать структуру проекта 
-реализовать другое чтение конфигов(например:
```go 
type Config struct{
  DB *ConfigPostrges
  RateLimiter *ConfigRL
  Backends *ConfigBackends
  ...
}
```
-
