# Bookstore API
**Bookstore API** - бекенд, написанный на Go. Монолит реализует REST_API для управления магазинами, книгами и товаром на складе.

## Запуск
1. Создать `.env` в корне проекта
```dotenv
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=bookstores
```
2. Запустить
```bash
docker compose up --build -d 
```
3. Провести миграцию
```bash
goose -dir "internal/database/migrations" postgres "host=localhost port=<DB_PORT> user=<DB_USER> password=<DB_PASSWORD> dbname=<DB_NAME> sslmode=disable" up
```
API будет доступен по адресу `http://localhost:8080` (дефолтная конфигурация)

## API Эндпоинты
### `/stores`
| Метод    | Путь                  | Описание                             |
| -------- | --------------------- | ------------------------------------ |
| `POST`   | `/stores`             | Создать новый магазин.               |
| `GET`    | `/stores`             | Получить список всех магазинов.      |
| `GET`    | `/stores/{storeUUID}` | Получить один магазин по UUID.       |
| `PUT`    | `/stores/{storeUUID}` | Обновить информацию о магазине.      |
| `DELETE` | `/stores/{storeUUID}` | "Закрыть" магазин (мягкое удаление). |

### `/books`
| Метод  | Путь                           | Описание                                      |
| ------ | ------------------------------ | --------------------------------------------- |
| `POST` | `/books`                       | Создать новую книгу в глобальном каталоге.    |
| `GET`  | `/books`                       | Получить список всех книг.                    |
| `GET`  | `/books/{bookID}`              | Получить одну книгу по ее ID.                 |
| `GET`  | `/books/search`                | Поиск книг по названию/автору (`?q=...`).     |
| `GET`  | `/books/{bookID}/availability` | Посмотреть, в каких магазинах доступна книга. |

### `/skus`
| Метод  | Путь                                | Описание                               |
| ------ | ----------------------------------- | -------------------------------------- |
| `POST` | `/skus`                             | Создать SKU (добавить книгу на склад). |
| `GET`  | `/skus/{skuUUID}`                   | Получить информацию о SKU.             |
| `PUT`  | `/skus/{skuUUID}/price`             | Обновить цену SKU.                     |
| `POST` | `/skus/{skuUUID}/stock-adjustments` | Сделать корректировку остатков.        |

## DB
Можно ознакомиться в [директории миграций](/internal/database/migrations)

## Технологии
- Go 1.25
- Chi v5
- POstgreSQL 18
- sqlc
- Goose
- cleanenv для конфига
- slog с кастомным middleware

## Перспективы
- Покрыть тестами
- CI/CD
- Авторизация и аутентификация (она должна была быть (JWT), но из-за сроков была временно вырезана)
- Юзеры, покупки, скидки