# Neo4j Cheatsheet

Referensi cepat Cypher query untuk project Family Tree.

---

## 🔍 MATCH (Read)

```cypher
-- Semua nodes
MATCH (n) RETURN n

-- By label
MATCH (p:Person) RETURN p

-- By property
MATCH (p:Person {name: "Hanif"}) RETURN p

-- By condition
MATCH (p:Person)
WHERE p.gender = "male" AND p.isDeleted = false
RETURN p.name, p.birthDate

-- Limit & skip (pagination)
MATCH (p:Person) RETURN p ORDER BY p.name SKIP 0 LIMIT 10

-- Count
MATCH (p:Person) RETURN count(p)

-- Search contains (case-insensitive)
MATCH (p:Person)
WHERE toLower(p.name) CONTAINS "hanif"
RETURN p
```

---

## ✏️ CREATE

```cypher
-- Create node
CREATE (p:Person {id: "uuid", name: "Hanif", gender: "male"})
RETURN p

-- Create relationship
MATCH (a:Person {id: "uuid-1"})
MATCH (b:Person {id: "uuid-2"})
CREATE (a)-[:PARENT_OF]->(b)

-- MERGE (upsert — bikin kalau belum ada)
MERGE (p:Person {id: "uuid"})
ON CREATE SET p.name = "Hanif", p.createdAt = datetime()
ON MATCH SET p.updatedAt = datetime()
```

---

## 🔄 SET (Update)

```cypher
-- Update properties
MATCH (p:Person {id: "uuid"})
SET p.name = "Hanif Naufal", p.updatedAt = datetime()
RETURN p

-- Add property
MATCH (p:Person {id: "uuid"})
SET p.bio = "Cloud Engineer"

-- Remove property
MATCH (p:Person {id: "uuid"})
REMOVE p.bio

-- Conditional update (COALESCE = pakai nilai baru kalau ga null)
MATCH (p:Person {id: "uuid"})
SET p.name = COALESCE($name, p.name)
```

---

## 🗑️ DELETE

```cypher
-- Delete node (harus ga punya relationship)
MATCH (p:Person {id: "uuid"}) DELETE p

-- DETACH DELETE (hapus node + semua relasi)
MATCH (p:Person {id: "uuid"}) DETACH DELETE p

-- Delete semua kecuali label tertentu
MATCH (n) WHERE NOT n:User DETACH DELETE n

-- Delete relationship only
MATCH (a)-[r:PARENT_OF]->(b) DELETE r
```

---

## 🔗 RELATIONSHIPS

```cypher
-- Cari semua relasi dari 1 node
MATCH (p:Person {id: "uuid"})-[r]->(related)
RETURN type(r), related.name

-- Cari relasi 2 arah
MATCH (p:Person {id: "uuid"})-[r]-(related)
RETURN type(r), related.name

-- Filter by relationship type
MATCH (p:Person)-[:PARENT_OF]->(child:Person)
RETURN p.name AS parent, child.name AS child

-- Multiple hops (family chain)
MATCH (p:Person {name: "Ahmad"})-[:PARENT_OF*1..3]->(descendant)
RETURN descendant.name

-- Optional match (LEFT JOIN-nya Neo4j)
MATCH (p:Person)
OPTIONAL MATCH (p)-[r]->(related)
RETURN p.name, type(r), related.name
```

---

## 📅 TYPES & FUNCTIONS

```cypher
-- Date & time
date("1995-01-15")                   -- Date type
datetime("2026-03-23T00:00:00Z")     -- DateTime
datetime()                           -- Now

-- String
toLower("HANIF")             -- "hanif"
toUpper("hanif")             -- "HANIF"
trim(" hanif ")              -- "hanif"
replace("hello", "l", "L")  -- "heLLo"

-- Null handling
COALESCE(null, "default")   -- "default"
p.name IS NULL
p.name IS NOT NULL

-- Aggregation
count(n), sum(n.age), avg(n.age), max(n.age), min(n.age)
```

---

## 🏷️ INDEX (Performance)

```cypher
-- Buat index
CREATE INDEX person_id FOR (p:Person) ON (p.id)
CREATE INDEX user_email FOR (u:User) ON (u.email)

-- List indexes
SHOW INDEXES

-- Drop index
DROP INDEX person_id
```

---

## 🧪 Useful untuk Debug

```cypher
-- Lihat semua labels
CALL db.labels()

-- Lihat semua relationship types
CALL db.relationshipTypes()

-- Lihat semua property keys
CALL db.propertyKeys()

-- Node count per label
MATCH (n) RETURN labels(n), count(n)

-- Explain query plan
EXPLAIN MATCH (p:Person {id: "uuid"}) RETURN p

-- Profile (run + show stats)
PROFILE MATCH (p:Person {id: "uuid"}) RETURN p
```

---

## 🌳 Family Tree Specific

```cypher
-- Semua ancestors (ke atas)
MATCH (p:Person {name: "X"})<-[:PARENT_OF*]-(ancestor)
RETURN ancestor.name

-- Semua descendants (ke bawah)
MATCH (p:Person {name: "X"})-[:PARENT_OF*]->(descendant)
RETURN descendant.name

-- Sibling (punya parent yang sama)
MATCH (p:Person {name: "X"})<-[:PARENT_OF]-(parent)-[:PARENT_OF]->(sibling)
WHERE sibling.name <> "X"
RETURN sibling.name

-- Shortest path antara 2 orang
MATCH path = shortestPath(
  (a:Person {name: "A"})-[*]-(b:Person {name: "B"})
)
RETURN path
```

---

## 🔗 Resources

- [Neo4j Cypher Docs](https://neo4j.com/docs/cypher-manual/current/)
- [Neo4j Browser](http://localhost:7474)
