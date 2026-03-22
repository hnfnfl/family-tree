# 🌳 Keluarga Tree - Silsilah Keluarga API

**Backend API untuk tracking silsilah keluarga berbasis Graph Database (Neo4j)**

---

## 🚀 Features

- ✅ **Graph-based Data Model** - Neo4j untuk relasi kompleks (poligami, perceraian, step-parent)
- ✅ **Smart Relationship Tracking** - Context-aware auto-complete untuk tambah relasi
- ✅ **RBAC Multi-user** - ADMIN, EDITOR, VIEWER roles
- ✅ **JWT Authentication** - Secure API endpoints
- ✅ **RESTful API** - Go + Gin framework
- ✅ **Docker Ready** - Easy deployment dengan Docker Compose
- ✅ **Age-Inclusive UX** - Support user 14-50+ tahun

---

## 📁 Project Structure

```
keluarga-tree/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # JWT auth, CORS, logging
│   ├── models/              # Data models
│   └── repository/          # Neo4j repository layer
├── deploy/
│   ├── docker-compose.yml   # Docker Compose config
│   └── *.json               # Neo4j test data scripts
├── docs/
│   ├── SCHEMA.md            # Neo4j schema design
│   └── queries.md           # Cypher query examples
├── docker-compose.yml       # Main Docker Compose
├── Dockerfile               # Go API Docker image
├── go.mod                   # Go module dependencies
├── PRD.md                   # Product Requirements Document
└── README.md                # This file
```

---

## 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| **Database** | Neo4j 5.x (Graph DB) |
| **Backend** | Go 1.23 + Gin Framework |
| **Auth** | JWT (golang-jwt/jwt/v5) |
| **Validation** | go-playground/validator |
| **Deploy** | Docker + Docker Compose |
| **Frontend** | React + react-flow (TODO) |

---

## 🏃 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (for local development)

### 1. Clone Repository
```bash
git clone https://github.com/hnfnfl/keluarga-tree.git
cd keluarga-tree
```

### 2. Start with Docker Compose
```bash
docker-compose up -d
```

This will start:
- **Neo4j** on http://localhost:7474 (username: `neo4j`, password: `KeluargaTree2026!`)
- **Go API** on http://localhost:8080

### 3. Verify Health
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2026-03-22T10:30:00Z",
  "service": "keluarga-tree-api",
  "version": "1.0.0"
}
```

---

## 📚 API Documentation

### Authentication

#### Register
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "role": "VIEWER"
}
```

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "...",
  "expires_at": "2026-03-23T10:30:00Z",
  "user": {
    "user_id": "...",
    "email": "user@example.com",
    "role": "VIEWER"
  }
}
```

### Persons

#### Get All Persons
```bash
GET /api/v1/persons?limit=100&offset=0
Authorization: Bearer <token>
```

#### Get Person by ID
```bash
GET /api/v1/persons/:id
Authorization: Bearer <token>
```

#### Create Person
```bash
POST /api/v1/persons
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Hanif Naufal Ashari",
  "gender": "male",
  "birthDate": "1995-01-01",
  "title": "Bc.",
  "bio": "Cloud Engineer @ SRIN",
  "addressStreet": "Jl. Tebah Raya No.2",
  "addressCity": "Jakarta Selatan",
  "addressProvince": "DKI Jakarta",
  "addressCountry": "Indonesia",
  "phonePrimary": "+6285730457714",
  "phonePrimaryType": "whatsapp"
}
```

### Families

#### Get Family Tree
```bash
GET /api/v1/families/:id/tree
```

---

## 🧪 Test Data

Neo4j browser sudah include test data dengan **5 keluarga** dan berbagai edge cases:
- Poligami
- Perceraian & pernikahan ulang
- Step-parent relationships
- Deceased members
- Single parents
- Siblings
- 4 generations

**Access Neo4j Browser:** http://localhost:7474

**Test queries:** Lihat `docs/queries.md`

---

## 🔧 Configuration

Environment variables (via `.env` or Docker Compose):

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Environment (development/production) |
| `SERVER_PORT` | `8080` | API server port |
| `NEO4J_URI` | `bolt://localhost:7687` | Neo4j connection URI |
| `NEO4J_USERNAME` | `neo4j` | Neo4j username |
| `NEO4J_PASSWORD` | `KeluargaTree2026!` | Neo4j password |
| `NEO4J_DATABASE` | `neo4j` | Neo4j database name |
| `JWT_SECRET` | `your-secret-key...` | JWT signing secret |
| `JWT_EXPIRE_HOUR` | `24` | Token expiry in hours |

---

## 📊 Database Schema

Lihat dokumentasi lengkap di:
- **Schema Design:** `docs/SCHEMA.md`
- **Cypher Queries:** `docs/queries.md`
- **Product Requirements:** `PRD.md`

---

## 🚀 Development

### Local Development (without Docker)

```bash
# Install dependencies
go mod download

# Run the application
go run cmd/main.go
```

### Run Tests
```bash
go test ./...
```

### Build Binary
```bash
go build -o keluarga-tree ./cmd/main.go
```

---

## 📝 TODO (Phase 2)

- [ ] Complete auth implementation (register/login)
- [ ] Family CRUD endpoints
- [ ] Relationship management endpoints
- [ ] Smart auto-complete queries
- [ ] Export to CSV/PDF
- [ ] Frontend React app
- [ ] PWA support
- [ ] Email notifications (N8N + Brevo)

---

## 👥 Team

**Owner:** Hanif Naufal Ashari  
**GitHub:** [@hnfnfl](https://github.com/hnfnfl)  
**Created:** March 2026

---

## 📄 License

Private - All rights reserved

---

*Built with ❤️ for Indonesian families*
