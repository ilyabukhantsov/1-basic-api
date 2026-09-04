# Product Catalog API

A small REST API for a product catalog written in Go using only the standard
`net/http` router (no web framework). Data is stored in SQLite via GORM,
authentication uses JWT (HS256).

Features:

- Categories: list, create, update, delete
- Products: create, update, delete, list by category. A product belongs to one
  or more categories (many-to-many)
- Authentication: `POST /auth/login` returns a JWT; all mutating endpoints
  require `Authorization: Bearer <token>`
- All responses, including errors, are JSON
- Graceful shutdown on `SIGINT`/`SIGTERM`
- OpenAPI spec in [`api/swagger.yaml`](api/swagger.yaml)

## Requirements

- Go 1.22+ (uses `net/http` method routing and `r.PathValue`)
- CGO enabled (the SQLite driver is `mattn/go-sqlite3`), i.e. a C compiler
  such as `gcc` must be available

## Running

```bash
cp .env.example .env        # then set SECRET_KEY to a long random string
go run ./cmd
```

The server listens on `:8080` by default. On first start it creates `app.db`
and seeds an admin user (`admin` / `admin123`, configurable via env).

### Configuration

All settings are read from the environment (a `.env` file is loaded if present):

| Variable         | Default    | Description                              |
|------------------|------------|------------------------------------------|
| `SECRET_KEY`     | *required* | HMAC secret for signing JWTs             |
| `ADDR`           | `:8080`    | Listen address                           |
| `DB_PATH`        | `app.db`   | SQLite file path (`:memory:` also works) |
| `TOKEN_TTL`      | `15m`      | JWT lifetime (Go duration syntax)        |
| `ADMIN_USERNAME` | `admin`    | Seeded admin username                    |
| `ADMIN_PASSWORD` | `admin123` | Seeded admin password                    |

## API

| Method | Path                         | Auth | Description                     |
|--------|------------------------------|------|---------------------------------|
| GET    | `/health`                    | no   | Health check                    |
| POST   | `/auth/login`                | no   | Get a JWT                       |
| GET    | `/categories`                | no   | List categories                 |
| GET    | `/categories/{id}/products`  | no   | List products in a category     |
| POST   | `/categories`                | yes  | Create category                 |
| PUT    | `/categories/{id}`           | yes  | Update category                 |
| DELETE | `/categories/{id}`           | yes  | Delete category (detaches products) |
| POST   | `/products`                  | yes  | Create product                  |
| PUT    | `/products/{id}`             | yes  | Update product                  |
| DELETE | `/products/{id}`             | yes  | Delete product                  |

Error responses have the shape `{"error": "message"}` with status codes
`400` (validation), `401` (auth), `404` (not found), `409` (duplicate
name/code), `413` (body over 1 MiB), `500` (unexpected).

### Example session

```bash
TOKEN=$(curl -s -X POST localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

curl -s -X POST localhost:8080/categories \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Laptops"}'
# {"id":1,"name":"Laptops"}

curl -s -X POST localhost:8080/products \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"MacBook","code":"MB-1","price":1999.99,"categoryIds":[1]}'
# {"id":1,"name":"MacBook","code":"MB-1","price":1999.99,"categoryIds":[1]}

curl -s localhost:8080/categories/1/products
```

## Tests

```bash
go test ./... -cover
```

Handler tests run the full router against an in-memory SQLite database, so
they cover routing, auth middleware, validation, the data layer and JSON
encoding end to end.

## Project layout

```
cmd/          entry point, config from env, graceful shutdown
handler/      HTTP handlers, router, JSON helpers
middleware/   JWT auth middleware and context helpers
database/     GORM connection, migrations, seed, data-access functions
models/       GORM models
jwt/          token manager
api/          OpenAPI spec
```

Handlers receive `*gorm.DB` and the JWT manager as explicit dependencies
(closures) rather than package globals, which keeps them testable.

## Time spent (approximate)

| Part                                                   | Time   |
|--------------------------------------------------------|--------|
| Project setup, models, DB connection, migrations       | 1 h    |
| Authentication (bcrypt, JWT, middleware)               | 1.5 h  |
| Categories CRUD                                        | 1 h    |
| Products CRUD and many-to-many handling                | 1.5 h  |
| Validation, error handling, JSON helpers               | 1 h    |
| Tests                                                  | 2 h    |
| OpenAPI spec, README, graceful shutdown, config        | 1 h    |
| **Total**                                              | **9 h** |
