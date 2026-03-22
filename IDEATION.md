# Keluarga Tree - Project Ideation

**Created:** 2026-03-XX  
**Status:** 💡 Ideation Phase  
**Owner:** Hanif Naufal Ashari  

---

## 🎯 Vision

Aplikasi silsilah keluarga berbasis graph database untuk tracking 5 generasi keluarga dengan visualisasi tree interaktif, search relasi, dan RBAC multi-user.

---

## 📋 Requirements Summary

### Core Features
- ✅ CRUD anggota keluarga
- ✅ Tambah/edit relasi (married, parent-child, siblings)
- ✅ Support 5 generasi (kepala keluarga → cicit)
- ✅ Tree visualization interaktif
- ✅ Search & query relasi kompleks
- ✅ Multi-user dengan RBAC

### RBAC Model
| Role | Permissions |
|------|-------------|
| `ADMIN` | Full CRUD semua anggota keluarga |
| `EDITOR` | CRUD diri sendiri + descendents |
| `VIEWER` | Read-only tree & search |

---

## 🏗️ Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Frontend  │────▶│  Backend API │────▶│   Neo4j     │
│  (React)    │◀────│    (Go)      │◀────│  (Graph DB) │
│  Tree Viz   │     │  Auth + RBAC │     │  5 Generasi │
└─────────────┘     └──────────────┘     └─────────────┘
```

### Tech Stack
| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Database** | Neo4j 5.x | Graph native, Cypher query powerful untuk relasi kompleks |
| **Backend** | Go (Gin/Echo) | Familiar (Nip udah jago), performant, concurrent |
| **Frontend** | React + react-flow | Tree viz interaktif, easier learning curve than D3 |
| **Auth** | JWT + bcrypt | Standard, secure, stateless |
| **Deploy** | Docker Compose | Homelab ready, easy maintenance |

---

## 📊 Neo4j Schema Design

### Nodes

```cypher
(:Person {
  id: uuid,
  name: string,
  gender: 'male'|'female',
  birthDate: date,
  deathDate: date?, 
  photo: url?,
  bio: string?,
  createdAt: datetime,
  updatedAt: datetime
})

(:Family {
  id: uuid,
  familyName: string,
  headOfFamilyId: uuid,
  createdAt: datetime
})

(:User {
  id: uuid,
  email: string,
  passwordHash: string,
  role: 'ADMIN'|'EDITOR'|'VIEWER',
  personId: uuid?,  // Link ke Person kalau user adalah anggota keluarga
  createdAt: datetime
})
```

### Relationships

```cypher
(:Person)-[:MARRIED_TO {sinceDate: date}]->(:Person)
(:Person)-[:PARENT_OF]->(:Person)
(:Person)-[:CHILD_OF]->(:Person)
(:Person)-[:SIBLING_OF]->(:Person)
(:Person)-[:BELONGS_TO]->(:Family)
(:User)-[:CAN_EDIT {role: string, grantedAt: datetime}]->(:Family)
```

---

## 🔍 Key Cypher Queries (Planned)

### 1. Get Full Family Tree (5 Generasi)
```cypher
MATCH (head:Person)-[:BELONGS_TO]->(f:Family)
WHERE head.id = $headId
OPTIONAL MATCH path = (head)-[:PARENT_OF*0..4]->(descendant)
RETURN path, head, f
```

### 2. Find Relationship Between Two People
```cypher
MATCH path = shortestPath(
  (p1:Person)-[*..10]-(p2:Person)
)
WHERE p1.id = $person1Id AND p2.id = $person2Id
RETURN path
```

### 3. Search Ancestors
```cypher
MATCH (person:Person)<-[:PARENT_OF*]-(ancestor)
WHERE person.id = $personId
RETURN ancestor ORDER BY ancestor.birthDate
```

### 4. Get All Descendants
```cypher
MATCH (head:Person)-[:PARENT_OF*]->(descendant)
WHERE head.id = $headId
RETURN descendant
```

---

## 📦 API Endpoints (Draft)

### Auth
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh token

### Family
- `POST /api/families` - Create new family
- `GET /api/families/:id` - Get family details
- `GET /api/families/:id/tree` - Get family tree (5 gen)
- `PUT /api/families/:id` - Update family (ADMIN only)

### Persons
- `POST /api/persons` - Add family member
- `GET /api/persons/:id` - Get person details
- `PUT /api/persons/:id` - Update person
- `DELETE /api/persons/:id` - Remove person
- `POST /api/persons/:id/relate` - Add relationship

### Search & Query
- `GET /api/search?q=` - Search by name
- `GET /api/relasi?person1=&person2=` - Find relationship
- `GET /api/persons/:id/ancestors` - Get ancestors
- `GET /api/persons/:id/descendants` - Get descendants

---

## 🎨 Frontend Components (Planned)

```
src/
├── components/
│   ├── Tree/
│   │   ├── FamilyTree.jsx       # Main tree viz (react-flow)
│   │   ├── PersonNode.jsx       # Custom node component
│   │   └── TreeControls.jsx     # Zoom, pan, filter
│   ├── Forms/
│   │   ├── PersonForm.jsx       # Add/edit person
│   │   ├── RelationshipForm.jsx # Add relationship
│   │   └── FamilyForm.jsx       # Create family
│   ├── Search/
│   │   ├── SearchBar.jsx        # Global search
│   │   └── RelasiFinder.jsx     # Find relationship tool
│   └── Auth/
│       ├── Login.jsx
│       └── Register.jsx
├── pages/
│   ├── Dashboard.jsx
│   ├── FamilyTree.jsx
│   ├── Search.jsx
│   └── Settings.jsx
└── utils/
    ├── api.js                   # Axios instance
    └── neo4j-queries.js         # Cypher query helpers
```

---

## 🚀 Development Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Setup Neo4j Docker container
- [ ] Design & test Cypher queries
- [ ] Backend scaffold (Go + Gin)
- [ ] Auth system (JWT)

### Phase 2: Core API (Week 2-3)
- [ ] Family CRUD endpoints
- [ ] Person CRUD endpoints
- [ ] Relationship management
- [ ] RBAC middleware

### Phase 3: Frontend (Week 3-5)
- [ ] React app setup
- [ ] Tree visualization (react-flow)
- [ ] Forms (Person, Family, Relationship)
- [ ] Search & Relasi Finder

### Phase 4: Polish & Deploy (Week 5-6)
- [ ] Testing (unit + integration)
- [ ] Docker Compose setup
- [ ] Deploy to homelab
- [ ] Documentation

---

## 📝 Open Questions / Decisions Needed

1. **Foto storage:** Local filesystem atau S3-compatible (MinIO)?
2. **Email notifications:** Perlu untuk invite family members?
3. **Export feature:** PDF/PNG tree export?
4. **Mobile app:** Priority atau web-responsive aja dulu?
5. **Backup strategy:** Neo4j backup schedule?

---

## 🔗 References

- Neo4j Docs: https://neo4j.com/docs/
- Cypher Query Language: https://neo4j.com/docs/cypher-manual/current/
- react-flow: https://reactflow.dev/
- Go Gin Framework: https://gin-gonic.com/

---

## 📌 Next Steps

1. ✅ **PRD Creation** - Detail specs, user stories, acceptance criteria
2. ⏳ **Neo4j Setup** - Docker container di homelab
3. ⏳ **Backend Scaffold** - Go project structure
4. ⏳ **Frontend Prototype** - React + react-flow POC

---

*Last Updated: 2026-03-XX*
