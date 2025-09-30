# File Downloader Service

Веб-сервис для параллельного скачивания файлов с поддержкой graceful shutdown и recovery.

## Возможности
- Параллельное скачивание файлов через worker pool
- Сохранение состояния задач в папку `state/` (создается в корне проекта)
- Graceful shutdown с сохранением состояния
- Автоматическое восстановление незавершенных задач
- REST API для управления задачами

## API Endpoints

### Создание задачи скачивания
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://i.pinimg.com/1200x/75/71/69/757169d55a4567d6f0b3e2df423af3a0.jpg", "https://file-examples.com/wp-content/uploads/2017/10/file-sample_150kB.pdf"]}'
```

### Получение статуса задачи
```bash
curl http://localhost:8080/api/v1/tasks/{task_id}/status
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Запуск

### Через Task
```bash
# Установка Task
go install github.com/go-task/task/v3/cmd/task@latest

# Сборка
task build

# Запуск
task run

# Тесты
task test
```

### Прямой запуск
```bash
# Сборка
go build -o filedownloader cmd/main.go

# Запуск
./filedownloader
```

## Конфигурация
Сервис загружает конфигурацию из `config.yaml`:
```yaml
server:
  port: 8080

worker:
  count: 3

logging:
  level: info
  format: json
  debug_mode: false
```

Переменные окружения переопределяют YAML:
- `SERVER_PORT` - порт сервера
- `WORKER_COUNT` - количество воркеров
- `LOG_LEVEL` - уровень логирования
- `LOG_FORMAT` - формат логов
- `DEBUG` - debug режим




