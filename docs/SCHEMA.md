# Neo4j Schema Design - Keluarga Tree

**Database:** Neo4j 5.x  
**Last Updated:** March 22, 2026  

---

## 📋 Overview

Schema ini menggunakan **flattened properties** untuk MVP (address & phone sebagai separate fields). Phase 2 akan migrate ke embedded maps/arrays untuk lebih flexibility.

---

## 👤 Person Node

### MVP Schema (Flattened Properties)

```cypher
(:Person {
  // Core fields (required)
  id: uuid,                          // randomUUID()
  name: string,
  gender: 'male'|'female'|'other',
  birthDate: date,                   // YYYY-MM-DD
  isDeleted: bool,                   // Soft delete flag
  createdAt: datetime,
  updatedAt: datetime
  
  // Optional fields
  deathDate: date?,                  // null = masih hidup
  title: string?,                    // Haji, Hj, Dr, Ir, etc.
  bio: string?,
  
  // Address (flattened for MVP)
  addressStreet: string?,
  addressNeighborhood: string?,      // RT/RW
  addressCity: string?,
  addressProvince: string?,
  addressPostalCode: string?,
  addressCountry: string?,
  
  // Phone (primary only for MVP)
  phonePrimary: string?,             // E.164 format: +62...
  phonePrimaryType: 'whatsapp'|'mobile'|'home'|'work',
  phoneVerified: bool?,              // Verified via OTP
  
  // Phase 2
  photoUrl: url?
})
```

### Phase 2 Schema (Embedded Objects)

```cypher
(:Person {
  // Core fields
  id: uuid,
  name: string,
  gender: 'male'|'female'|'other',
  birthDate: date,
  deathDate: date?,
  isDeleted: bool,
  createdAt: datetime,
  updatedAt: datetime
  
  // Optional
  title: string?,
  bio: string?,
  
  // Embedded address object
  address: {
    street: string,
    neighborhood: string?,
    city: string,
    province: string,
    postalCode: string?,
    country: string,
    coordinates: {lat: float, lng: float}?,
    isPrimary: bool,
    validFrom: date?,
    validUntil: date?
  }?,
  
  // Array of phone numbers
  phoneNumbers: [
    {
      number: string,                // E.164 format
      type: 'whatsapp'|'mobile'|'home'|'work',
      label: string?,
      isPrimary: bool,
      isVerified: bool,
      visibility: 'public'|'family_only'|'admin_only'|'private',
      addedAt: datetime
    }
  ],
  
  photoUrl: url?
})
```

---

## 👨‍👩‍👧‍👦 Family Node

```cypher
(:Family {
  id: uuid,                          // randomUUID()
  familyName: string,
  headOfFamilyId: uuid,              // Link to Person.id
  createdAt: datetime
})
```

---

## 🔐 User Node (Authentication)

```cypher
(:User {
  id: uuid,                          // randomUUID()
  email: string,                     // Unique, login credential
  passwordHash: string,              // bcrypt (cost: 12)
  role: 'ADMIN'|'EDITOR'|'VIEWER',
  personId: uuid?,                   // Link to Person.id (optional)
  lastLoginAt: datetime?,
  createdAt: datetime
})
```

---

## 💑 Relationships

### MARRIED_TO

```cypher
(:Person)-[:MARRIED_TO {
  startDate: date,
  endDate: date?,                    // null = still married
  endReason: 'divorce'|'death'?,
  isCurrent: bool,                   // true = active marriage
  order: number                      // 1st, 2nd, 3rd, 4th wife
}]->(:Person)
```

### PARENT_OF / CHILD_OF

```cypher
(:Person)-[:PARENT_OF]->(:Person)
(:Person)-[:CHILD_OF]->(:Person)
```

### SIBLING_OF

```cypher
(:Person)-[:SIBLING_OF]->(:Person)
```

### STEP_PARENT_OF

```cypher
(:Person)-[:STEP_PARENT_OF {
  sinceDate: date
}]->(:Person)
```

### BELONGS_TO

```cypher
(:Person)-[:BELONGS_TO]->(:Family)
```

### CAN_EDIT (User → Family)

```cypher
(:User)-[:CAN_EDIT {
  role: 'ADMIN'|'EDITOR'|'VIEWER',
  grantedAt: datetime
}]->(:Family)
```

---

## 📊 Indexes & Constraints

### Unique Constraints

```cypher
CREATE CONSTRAINT person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE;
CREATE CONSTRAINT family_id IF NOT EXISTS FOR (f:Family) REQUIRE f.id IS UNIQUE;
CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE;
CREATE CONSTRAINT user_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE;
```

### Indexes

```cypher
CREATE INDEX person_name IF NOT EXISTS FOR (p:Person) ON (p.name);
CREATE INDEX person_birth_date IF NOT EXISTS FOR (p:Person) ON (p.birthDate);
CREATE INDEX person_city IF NOT EXISTS FOR (p:Person) ON (p.addressCity);
CREATE INDEX person_phone IF NOT EXISTS FOR (p:Person) ON (p.phonePrimary);
CREATE INDEX family_name IF NOT EXISTS FOR (f:Family) ON (f.familyName);
```

---

## 🔄 Migration Path (MVP → Phase 2)

### Migrate Address (Flattened → Embedded)

```cypher
MATCH (p:Person)
WHERE p.addressStreet IS NOT NULL
SET p.address = {
  street: p.addressStreet,
  neighborhood: p.addressNeighborhood,
  city: p.addressCity,
  province: p.addressProvince,
  postalCode: p.addressPostalCode,
  country: p.addressCountry,
  isPrimary: true
}
REMOVE p.addressStreet, p.addressNeighborhood, p.addressCity, 
       p.addressProvince, p.addressPostalCode, p.addressCountry
RETURN count(p) as migratedCount;
```

### Migrate Phone (Flattened → Array)

```cypher
MATCH (p:Person)
WHERE p.phonePrimary IS NOT NULL
SET p.phoneNumbers = [{
  number: p.phonePrimary,
  type: p.phonePrimaryType,
  label: 'Pribadi',
  isPrimary: true,
  isVerified: p.phoneVerified,
  visibility: 'family_only',
  addedAt: p.createdAt
}]
REMOVE p.phonePrimary, p.phonePrimaryType, p.phoneVerified
RETURN count(p) as migratedCount;
```

---

## 📝 Example Data

### Complete Person (MVP)

```cypher
CREATE (p:Person {
  id: randomUUID(),
  name: 'Hanif Naufal Ashari',
  gender: 'male',
  birthDate: date('1995-01-01'),
  title: 'Bc.',
  bio: 'Cloud Engineer @ SRIN',
  addressStreet: 'Jl. Tebah Raya No.2',
  addressNeighborhood: 'Kebayoran Baru',
  addressCity: 'Jakarta Selatan',
  addressProvince: 'DKI Jakarta',
  addressPostalCode: '12180',
  addressCountry: 'Indonesia',
  phonePrimary: '+6285730457714',
  phonePrimaryType: 'whatsapp',
  phoneVerified: true,
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})
RETURN p;
```

### Complete Person (Phase 2)

```cypher
CREATE (p:Person {
  id: randomUUID(),
  name: 'Hanif Naufal Ashari',
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
  phoneNumbers: [{
    number: '+6285730457714',
    type: 'whatsapp',
    label: 'Pribadi',
    isPrimary: true,
    isVerified: true,
    visibility: 'family_only',
    addedAt: datetime()
  }],
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})
RETURN p;
```

---

*Last Updated: March 22, 2026*
