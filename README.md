# ДЗ№1 по курсу "Безопасность интернет приложений"
Запуск:
- ./gen_ca.sh
- docker-compose up
- go run main.go

Функционал:
- http прокси
- https прокси
- сохранение запросов в базу данных
- повтор запросов
- анализ параметров запроса на наличие уязвимости "command injection"

Ручки:
- requests - вывод всех запросов, сохраненных в БД
- request/id - вывод запроса
- repead/id - повтор запроса
- scan/id - анализ запроса на наличие уязвимости
