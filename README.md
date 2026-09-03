# mythings

[Banner](documentation/banner.png)

mythings is a personal catalog application for keeping track of the things you own.

The project lets you create a structured database of personal items, attach tags and images, store purchase information, and quickly find items by name, tags, or price range.

## Features

- create, view, edit, and delete personal items;
- search items by name;
- create and manage tags;
- filter items by one or more tags;
- filter items by minimum and maximum price;
- store short and full descriptions;
- store purchase date and price information;
- RUB and USD price support with a stored USD exchange rate;
- upload item images in JPEG, PNG, WEBP, and GIF formats;
- REST API for items, tags, and uploads;

## Getting started

### Run with Docker Compose

Clone the repository:

```bash
git clone https://github.com/Mimist-Illusionard/mythings
cd mythings
```

Build and start the application:

```bash
docker compose up --build
```

After startup, open:

```text
http://localhost:8080
```

Docker Compose starts both the application and PostgreSQL. Uploaded images and database data are stored in persistent Docker volumes.

### Run locally

Requirements:

- Go 1.26.1;
- PostgreSQL.

Clone the repository:

```bash
git clone https://github.com/Mimist-Illusionard/mythings
cd mythings
```

Create an environment file:

```bash
cp .env.example .env
```

Default configuration:

```env
DB_NAME=mythings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
```

Install dependencies:

```bash
go mod tidy
```

Build the application:

```bash
go build -o mythings ./cmd/mythings
```

Run it:

```bash
./mythings
```

The application automatically applies embedded PostgreSQL migrations on startup.

## Usage

### Add an item

Open the web interface and click **New item**.

An item can contain:

- name;
- short description;
- full description;
- purchase date;
- price;
- currency;
- USD exchange rate;
- image;
- tags.

### Search and filter

The catalog can be filtered by:

- partial item name;
- minimum price;
- maximum price;
- one or more tags.

When several tags are selected, an item must contain every selected tag to match the filter.

For price filtering, USD prices are converted to RUB using the exchange rate stored with the item.

### Tags

Tags can be created independently and attached to any number of items.

Each item can have multiple tags, which makes it possible to organize the same collection in different ways without introducing a fixed folder or category hierarchy.

Examples:

```text
Electronics
Books
Work
Travel
Collectibles
For sale
```

## Command-line flags

| Flag | Description |
|------|-------------|
| `-env <path>` | Path to the environment file. Default: `.env` |
| `-port <port>` | HTTP server port. Default: `8080` |

Example:

```bash
./mythings -env ./config/local.env -port 8090
```

## REST API

### Items

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/items` | List and filter items |
| `GET` | `/items/{id}` | Get an item by ID |
| `POST` | `/items` | Create an item |
| `PUT` | `/items/{id}` | Update an item |
| `DELETE` | `/items/{id}` | Delete an item |
| `POST` | `/items/{id}/tags/{tagID}` | Attach a tag to an item |
| `DELETE` | `/items/{id}/tags/{tagID}` | Remove a tag from an item |

Supported query parameters for `GET /items`:

| Parameter | Description |
|-----------|-------------|
| `name` | Search by partial item name |
| `tag` | Filter by tag name. Can be specified multiple times |
| `min_price` | Minimum price in RUB |
| `max_price` | Maximum price in RUB |
| `limit` | Number of results, from 1 to 100. Default: `50` |
| `offset` | Pagination offset. Default: `0` |

Example:

```bash
curl "http://localhost:8080/items?name=watch&tag=Electronics&min_price=1000&limit=20"
```

### Tags

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/tags` | List tags |
| `POST` | `/tags` | Create a tag |
| `PUT` | `/tags/{id}` | Rename a tag |
| `DELETE` | `/tags/{name}` | Delete a tag by name |

Tags can also be searched by partial name:

```bash
curl "http://localhost:8080/tags?name=book"
```

### Uploads

Upload an image using `multipart/form-data` with the `image` field:

```bash
curl -X POST \
  -F "image=@./photo.png" \
  http://localhost:8080/uploads
```

Supported image formats:

- JPEG;
- PNG;
- WEBP;
- GIF.

Maximum image size is 10 MB.

A successful upload returns the URL that can be stored in an item:

```json
{
  "url": "/uploads/0f8369005db1256af2b0de33.png"
}
```

## Example item

```json
{
  "name": "Mechanical keyboard",
  "short_description": "75% wireless keyboard",
  "description": "Personal keyboard used for work and programming.",
  "image_url": "/uploads/0f8369005db1256af2b0de33.png",
  "price": 120,
  "price_currency": "USD",
  "usd_exchange_rate": 81.5,
  "purchased_at": "2026-08-15",
  "attributes": {
    "layout": "75%",
    "connection": "Bluetooth"
  }
}
```

## Project structure

```text
mythings/
├── cmd/mythings/                     # application entry point
├── config/                           # environment configuration
├── internal/
│   ├── app/                          # application startup and HTTP server
│   ├── domain/
│   │   ├── models/                   # item and tag domain models
│   │   └── ports/repository/         # repository interfaces
│   ├── infrastructure/postgres/      # PostgreSQL repositories and migrations
│   └── interfaces/http/handlers/     # REST API handlers
├── web/
│   ├── static/                       # frontend JavaScript and styles
│   ├── uploads/                      # uploaded item images
│   └── index.html                    # web interface
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## Tech stack

- Go;
- PostgreSQL;
- `pgx`;
- `gorilla/mux`;
- `golang-migrate`;
- HTML, CSS, and vanilla JavaScript;
- Docker and Docker Compose.
