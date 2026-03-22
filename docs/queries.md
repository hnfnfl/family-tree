# Neo4j Cypher Queries - Keluarga Tree

**Database:** Neo4j 5.x  
**Auth:** `neo4j` / `KeluargaTree2026!`  
**Bolt URL:** bolt://192.168.18.65:7687  
**Browser URL:** http://192.168.18.65:7474  

---

## 📋 Schema Setup

### 1. Create Constraints & Indexes

```cypher
// Unique constraints
CREATE CONSTRAINT person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE CONSTRAINT family_id IF NOT EXISTS FOR (f:Family) REQUIRE f.id IS UNIQUE;
CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE;
CREATE CONSTRAINT user_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE;

// Indexes for performance
CREATE INDEX person_name IF NOT EXISTS FOR (p:Person) ON (p.name);
CREATE INDEX person_birth_date IF NOT EXISTS FOR (p:Person) ON (p.birthDate);
CREATE INDEX family_name IF NOT EXISTS FOR (f:Family) ON (f.familyName);
```

---

## 👤 Person CRUD

### Create Person (Complete)

```cypher
CREATE (p:Person {
  id: randomUUID(),
  name: 'Hanif Naufal Ashari',
  gender: 'male',
  birthDate: date('1995-01-01'),
  deathDate: null,
  title: 'Bc.',
  bio: 'Cloud Engineer @ SRIN',
  address: {
    street: 'Jl. Tebah Raya No.2',
    neighborhood: 'Kebayoran Baru',
    city: 'Jakarta Selatan',
    province: 'DKI Jakarta',
    postalCode: '12180',
    country: 'Indonesia',
    isPrimary: true
  },
  phoneNumbers: [
    {
      number: '+6285730457714',
      type: 'whatsapp',
      label: 'Pribadi',
      isPrimary: true,
      isVerified: false,
      visibility: 'family_only',
      addedAt: datetime()
    }
  ],
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})
RETURN p;
```

### Create Person (Minimal)

```cypher
CREATE (p:Person {
  id: randomUUID(),
  name: 'Siti Aminah',
  gender: 'female',
  birthDate: date('1990-01-01'),
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})
RETURN p;
```

### Get Person by ID

```cypher
MATCH (p:Person {id: $personId})
WHERE p.isDeleted = false
RETURN p;
```

### Update Person

```cypher
MATCH (p:Person {id: $personId})
SET p.name = $name,
    p.bio = $bio,
    p.updatedAt = datetime()
RETURN p;
```

### Soft Delete Person

```cypher
MATCH (p:Person {id: $personId})
SET p.isDeleted = true,
    p.updatedAt = datetime()
RETURN p;
```

---

## 👨‍👩‍👧‍👦 Family CRUD

### Create Family

```cypher
CREATE (f:Family {
  id: randomUUID(),
  familyName: 'Keluarga Besar Ashari',
  headOfFamilyId: $headOfFamilyId,
  createdAt: datetime()
})
RETURN f;
```

### Get Family Tree (5 Generations)

```cypher
MATCH (head:Person {id: $headId})-[:BELONGS_TO]->(f:Family)
OPTIONAL MATCH path = (head)-[:PARENT_OF*0..4]->(descendant)
WHERE descendant.isDeleted = false
RETURN path, head, f
ORDER BY descendant.birthDate;
```

### Get Family Members

```cypher
MATCH (p:Person)-[:BELONGS_TO]->(f:Family {id: $familyId})
WHERE p.isDeleted = false
RETURN p
ORDER BY p.birthDate;
```

---

## 💑 Marriage Relationships

### Add Marriage (First Wife)

```cypher
MATCH (husband:Person {id: $husbandId})
MATCH (wife:Person {id: $wifeId})
MERGE (husband)-[r:MARRIED_TO {
  startDate: date('2020-01-01'),
  endDate: null,
  endReason: null,
  isCurrent: true,
  order: 1
}]->(wife)
RETURN r;
```

### Add Second Marriage (Polygamy)

```cypher
MATCH (husband:Person {id: $husbandId})
MATCH (wife2:Person {id: $wife2Id})
MERGE (husband)-[r:MARRIED_TO {
  startDate: date('2023-01-01'),
  endDate: null,
  endReason: null,
  isCurrent: true,
  order: 2
}]->(wife2)
RETURN r;
```

### End Marriage (Divorce)

```cypher
MATCH (p1:Person {id: $person1Id})-[r:MARRIED_TO {isCurrent: true}]->(p2:Person {id: $person2Id})
SET r.endDate = date(),
    r.endReason = 'divorce',
    r.isCurrent = false
RETURN r;
```

### End Marriage (Death - Auto)

```cypher
// When a person dies, auto-end all their current marriages
MATCH (deceased:Person {id: $deceasedId})
MATCH (deceased)-[r:MARRIED_TO {isCurrent: true}]-(spouse)
SET r.endDate = deceased.deathDate,
    r.endReason = 'death',
    r.isCurrent = false
RETURN r, spouse;
```

### Get Current Spouse(s)

```cypher
MATCH (p:Person {id: $personId})-[r:MARRIED_TO {isCurrent: true}]-(spouse)
RETURN spouse, r.startDate, r.order
ORDER BY r.order;
```

### Get Marriage History

```cypher
MATCH (p:Person {id: $personId})-[r:MARRIED_TO]-(spouse)
WHERE p.isDeleted = false AND spouse.isDeleted = false
RETURN spouse, r.startDate, r.endDate, r.endReason, r.isCurrent, r.order
ORDER BY r.order;
```

---

## 👶 Parent-Child Relationships

### Add Parent-Child

```cypher
MATCH (parent:Person {id: $parentId})
MATCH (child:Person {id: $childId})
MERGE (parent)-[:PARENT_OF]->(child)
MERGE (child)-[:CHILD_OF]->(parent)
RETURN parent, child;
```

### Get All Children

```cypher
MATCH (parent:Person {id: $parentId})-[:PARENT_OF]->(child:Person)
WHERE child.isDeleted = false
RETURN child
ORDER BY child.birthDate;
```

### Get All Parents

```cypher
MATCH (child:Person {id: $childId})<-[:PARENT_OF]-(parent:Person)
WHERE parent.isDeleted = false
RETURN parent;
```

### Get Ancestors (All Generations)

```cypher
MATCH (person:Person {id: $personId})<-[:PARENT_OF*]-(ancestor:Person)
WHERE ancestor.isDeleted = false
RETURN ancestor, size((person)<-[:PARENT_OF*]-(ancestor)) as generation
ORDER BY generation DESC, ancestor.birthDate;
```

### Get Descendants (5 Generations)

```cypher
MATCH (head:Person {id: $headId})-[:PARENT_OF*0..4]->(descendant:Person)
WHERE descendant.isDeleted = false
RETURN descendant, size((head)-[:PARENT_OF*]->(descendant)) as generation
ORDER BY generation, descendant.birthDate;
```

---

## 👫 Sibling Relationships

### Get Siblings (Same Parents)

```cypher
MATCH (person:Person {id: $personId})<-[:PARENT_OF]-(parent)-[:PARENT_OF]-(sibling:Person)
WHERE sibling.id <> person.id AND sibling.isDeleted = false
RETURN DISTINCT sibling
ORDER BY sibling.birthDate;
```

### Add Sibling Relationship (Manual)

```cypher
MATCH (sibling1:Person {id: $sibling1Id})
MATCH (sibling2:Person {id: $sibling2Id})
MERGE (sibling1)-[:SIBLING_OF]-(sibling2)
RETURN sibling1, sibling2;
```

---

## 👨‍👧 Step-Parent Relationships

### Add Step-Parent

```cypher
MATCH (stepParent:Person {id: $stepParentId})
MATCH (stepChild:Person {id: $stepChildId})
MATCH (bioParent:Person)-[:PARENT_OF]->(stepChild)
MATCH (stepParent)-[:MARRIED_TO {isCurrent: true}]-(bioParent)
MERGE (stepParent)-[r:STEP_PARENT_OF {sinceDate: date()}]->(stepChild)
RETURN r, stepParent, stepChild;
```

### Get Step-Parents

```cypher
MATCH (child:Person {id: $childId})<-[:STEP_PARENT_OF]-(stepParent:Person)
WHERE stepParent.isDeleted = false
RETURN stepParent;
```

### Get Step-Children

```cypher
MATCH (stepParent:Person {id: $stepParentId})-[:STEP_PARENT_OF]->(stepChild:Person)
WHERE stepChild.isDeleted = false
RETURN stepChild;
```

---

## 🔍 Smart Auto-Complete Queries

### Find Valid Parent Candidate

```cypher
// Exclude: self, descendants, siblings
MATCH (person:Person {id: $personId})
MATCH (candidate:Person)
WHERE candidate.id <> person.id
AND candidate.isDeleted = false
AND NOT (person)-[:PARENT_OF*]->(candidate)  // Not descendant
AND NOT (candidate)-[:PARENT_OF*]->(person)  // Not ancestor (will be added as parent)
RETURN candidate.name, candidate.birthDate, candidate.gender
ORDER BY candidate.birthDate ASC  // Older first
LIMIT 10;
```

### Find Valid Child Candidate

```cypher
// Exclude: self, ancestors, siblings
MATCH (person:Person {id: $personId})
MATCH (candidate:Person)
WHERE candidate.id <> person.id
AND candidate.isDeleted = false
AND NOT (candidate)-[:PARENT_OF*]->(person)  // Not ancestor
AND NOT (person)-[:PARENT_OF*]->(candidate)  // Not descendant (will be added as child)
RETURN candidate.name, candidate.birthDate, candidate.gender
ORDER BY candidate.birthDate DESC  // Younger first
LIMIT 10;
```

### Find Valid Spouse Candidate

```cypher
// Exclude: self, ancestors, descendants, siblings, current spouses
MATCH (person:Person {id: $personId})
MATCH (candidate:Person)
WHERE candidate.id <> person.id
AND candidate.isDeleted = false
AND NOT (candidate)-[:PARENT_OF*]->(person)  // Not ancestor
AND NOT (person)-[:PARENT_OF*]->(candidate)  // Not descendant
AND NOT EXISTS((candidate)<-[:PARENT_OF]-(parent)-[:PARENT_OF]->(person))  // Not sibling
AND NOT (candidate)-[:MARRIED_TO {isCurrent: true}]-(person)  // Not current spouse
RETURN candidate.name, candidate.birthDate, candidate.gender
ORDER BY candidate.name
LIMIT 10;
```

### Find Valid Sister Candidate (Smart Filter)

```cypher
// Find younger female siblings only
MATCH (person:Person {id: $personId})<-[:PARENT_OF]-(parent)-[:PARENT_OF]-(sibling:Person)
WHERE sibling.id <> person.id
AND sibling.isDeleted = false
AND sibling.gender = 'female'
AND sibling.birthDate > person.birthDate  // Younger (adek)
RETURN sibling.name, sibling.birthDate
ORDER BY sibling.birthDate DESC;
```

---

## 🔗 Relationship Finder

### Find Relationship Between Two People

```cypher
MATCH path = shortestPath(
  (p1:Person {id: $person1Id})-[*..10]-(p2:Person {id: $person2Id})
)
WHERE p1.isDeleted = false AND p2.isDeleted = false
RETURN path;
```

### Get Relationship Path with Types

```cypher
MATCH path = shortestPath(
  (p1:Person {id: $person1Id})-[*..10]-(p2:Person {id: $person2Id})
)
WHERE p1.isDeleted = false AND p2.isDeleted = false
RETURN 
  [rel IN relationships(path) | type(rel)] as relationshipTypes,
  [node IN nodes(path) | node.name] as personNames;
```

---

## 📊 Search & Export

### Search & Filter

### Search by Name (Fuzzy)

```cypher
// Simple contains search
MATCH (p:Person)
WHERE toLower(p.name) CONTAINS toLower($searchTerm)
AND p.isDeleted = false
RETURN p.name, p.birthDate, p.gender
ORDER BY p.name
LIMIT 20;
```

### Search by Phone Number

```cypher
// Search by phone/WA number
MATCH (p:Person)
WHERE ANY(num IN p.phoneNumbers WHERE num.number CONTAINS $phoneNumber)
AND p.isDeleted = false
RETURN p.name, p.phoneNumbers, p.address.city
ORDER BY p.name;
```

### Search by City

```cypher
// Find all family members in a city
MATCH (p:Person)
WHERE p.address.city = $city
AND p.isDeleted = false
RETURN p.name, p.address, 
       [num IN p.phoneNumbers WHERE num.isPrimary = true | num.number][0] as primaryPhone
ORDER BY p.name;
```

### Get Family Phone Directory

```cypher
// Export phone directory for WhatsApp group
MATCH (p:Person)-[:BELONGS_TO]->(f:Family {id: $familyId})
WHERE p.isDeleted = false
AND p.phoneNumbers IS NOT NULL
AND size(p.phoneNumbers) > 0
RETURN 
  p.name,
  [num IN p.phoneNumbers WHERE num.type = 'whatsapp' AND num.isPrimary = true | num.number][0] as waNumber,
  p.address.city
ORDER BY p.name;
```

### Get Family Address List

```cypher
// Export address list for sending parcels/invitations
MATCH (p:Person)-[:BELONGS_TO]->(f:Family {id: $familyId})
WHERE p.isDeleted = false
AND p.address IS NOT NULL
RETURN 
  p.name,
  p.address.street,
  p.address.neighborhood,
  p.address.city,
  p.address.province,
  p.address.postalCode,
  p.address.country
ORDER BY p.address.city, p.name;
```

### Get WhatsApp Quick Links

```cypher
// Generate wa.me links for all family members
MATCH (p:Person)-[:BELONGS_TO]->(f:Family {id: $familyId})
WHERE p.isDeleted = false
AND p.phoneNumbers IS NOT NULL
WITH p, [num IN p.phoneNumbers WHERE num.type = 'whatsapp' AND num.isPrimary = true | num.number][0] as waNumber
WHERE waNumber IS NOT NULL
RETURN 
  p.name,
  waNumber,
  'https://wa.me/' + replace(waNumber, '+', '') as waLink
ORDER BY p.name;
```

### Export All Family Members (CSV Format)

```cypher
MATCH (p:Person)-[:BELONGS_TO]->(f:Family {id: $familyId})
WHERE p.isDeleted = false
RETURN 
  p.id,
  p.name,
  p.gender,
  p.birthDate,
  p.deathDate,
  p.bio,
  f.familyName
ORDER BY p.birthDate;
```

### Get Generation Count

```cypher
MATCH (head:Person {id: $headId})-[:PARENT_OF*0..4]->(descendant:Person)
WHERE descendant.isDeleted = false
WITH 
  size((head)-[:PARENT_OF*]->(descendant)) as generation,
  count(descendant) as count
RETURN 
  CASE generation
    WHEN 0 THEN 'Head (Gen 1)'
    WHEN 1 THEN 'Children (Gen 2)'
    WHEN 2 THEN 'Grandchildren (Gen 3)'
    WHEN 3 THEN 'Great-Grandchildren (Gen 4)'
    WHEN 4 THEN 'Cicit (Gen 5)'
  END as generationName,
  count
ORDER BY generation;
```

---

## ⚠️ Validation Queries

### Check for Circular Parent Relationships

```cypher
// Should return 0 results
MATCH (p:Person)-[:PARENT_OF*]->(p)
RETURN p;
```

### Check for Date Inconsistencies

```cypher
// Child born before parent
MATCH (parent:Person)-[:PARENT_OF]->(child:Person)
WHERE child.birthDate < parent.birthDate + duration({years: 10})
RETURN parent.name, parent.birthDate, child.name, child.birthDate;

// Child born after parent death (+9 months)
MATCH (parent:Person)-[:PARENT_OF]->(child:Person)
WHERE parent.deathDate IS NOT NULL
AND child.birthDate > parent.deathDate + duration({months: 9})
RETURN parent.name, parent.deathDate, child.name, child.birthDate;

// Death before birth
MATCH (p:Person)
WHERE p.deathDate < p.birthDate
RETURN p.name, p.birthDate, p.deathDate;
```

### Check for Duplicate Marriages

```cypher
// Same couple married multiple times (active)
MATCH (p1:Person)-[r1:MARRIED_TO {isCurrent: true}]->(p2:Person)<-[r2:MARRIED_TO {isCurrent: true}]-(p1)
WHERE id(r1) <> id(r2)
RETURN p1.name, p2.name, count(r1) as duplicateCount;
```

---

## 🧪 Test Data Setup (Complete)

```cypher
// Create Head of Family
CREATE (head:Person {
  id: randomUUID(),
  name: 'Ahmad Ashari',
  gender: 'male',
  birthDate: date('1960-01-01'),
  title: 'Haji',
  bio: 'Kepala Keluarga',
  address: {
    street: 'Jl. Tebah Raya No.2',
    neighborhood: 'Kebayoran Baru',
    city: 'Jakarta Selatan',
    province: 'DKI Jakarta',
    postalCode: '12180',
    country: 'Indonesia',
    isPrimary: true
  },
  phoneNumbers: [
    {
      number: '+6281234567890',
      type: 'whatsapp',
      label: 'Pribadi',
      isPrimary: true,
      isVerified: false,
      visibility: 'family_only',
      addedAt: datetime()
    }
  ],
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})

// Create Family
CREATE (f:Family {
  id: randomUUID(),
  familyName: 'Keluarga Besar Ashari',
  headOfFamilyId: head.id,
  createdAt: datetime()
})

// Link Head to Family
MERGE (head)-[:BELONGS_TO]->(f)

// Create Wife 1
CREATE (wife1:Person {
  id: randomUUID(),
  name: 'Siti Aminah',
  gender: 'female',
  birthDate: date('1965-01-01'),
  title: 'Hj.',
  address: {
    street: 'Jl. Tebah Raya No.2',
    neighborhood: 'Kebayoran Baru',
    city: 'Jakarta Selatan',
    province: 'DKI Jakarta',
    postalCode: '12180',
    country: 'Indonesia',
    isPrimary: true
  },
  phoneNumbers: [
    {
      number: '+6281234567891',
      type: 'whatsapp',
      label: 'Pribadi',
      isPrimary: true,
      isVerified: false,
      visibility: 'family_only',
      addedAt: datetime()
    }
  ],
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})

// Create Marriage (First)
MERGE (head)-[:MARRIED_TO {
  startDate: date('1985-01-01'),
  isCurrent: true,
  order: 1
}]->(wife1)

// Create Child 1
CREATE (child1:Person {
  id: randomUUID(),
  name: 'Hanif Ashari',
  gender: 'male',
  birthDate: date('1995-01-01'),
  title: 'Bc.',
  bio: 'Cloud Engineer @ SRIN',
  address: {
    street: 'Jl. Tebah Raya No.2',
    neighborhood: 'Kebayoran Baru',
    city: 'Jakarta Selatan',
    province: 'DKI Jakarta',
    postalCode: '12180',
    country: 'Indonesia',
    isPrimary: true
  },
  phoneNumbers: [
    {
      number: '+6285730457714',
      type: 'whatsapp',
      label: 'Pribadi',
      isPrimary: true,
      isVerified: true,
      visibility: 'family_only',
      addedAt: datetime()
    }
  ],
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})

// Link Child to Parents
MERGE (head)-[:PARENT_OF]->(child1)
MERGE (wife1)-[:PARENT_OF]->(child1)
MERGE (child1)-[:BELONGS_TO]->(f)

RETURN head, wife1, child1, f;
```

---

*Last Updated: March 22, 2026*
