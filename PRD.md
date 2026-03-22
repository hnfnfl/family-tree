# 📋 PRD - Keluarga Tree (Silsilah Keluarga)

**Version:** 1.1 (Age-Inclusive Update)  
**Created:** March 21, 2026  
**Updated:** March 21, 2026 (Age-specific UX)  
**Owner:** Hanif Naufal Ashari  
**Status:** ✅ Approved for Development  

---

## 1. Executive Summary

### Problem Statement
Keluarga besar Nip membutuhkan sistem terpusat untuk tracking silsilah keluarga hingga 5 generasi dengan kompleksitas relasi (poligami, perceraian, pernikahan ulang, anak tiri). Metode manual (catatan, Excel) tidak scalable dan rawan error.

### Solution Overview
Aplikasi web berbasis graph database (Neo4j) untuk tracking silsilah keluarga dengan:
- Visualisasi tree interaktif
- Smart relationship tracking (poligami, perceraian, step-parent)
- Context-aware auto-complete untuk tambah relasi
- RBAC multi-user (ADMIN/EDITOR/VIEWER)
- Export data (CSV/PDF list)

### Target Users
**Age Range:** 14-50+ tahun (Remaja → Eyang)

- **Kepala Keluarga** (ADMIN) - Manage seluruh family data
- **Orang Tua** (EDITOR) - Manage diri sendiri + descendants
- **Anak/Cucu** (VIEWER) - Lihat tree, search relasi
- **Tech Admin** (Nip) - Maintain infra

**Design Principle:** Inclusive UX untuk semua generasi (teen tech-savvy → senior gaptek)

---

## 2. User Personas

### Primary Personas

| Persona | Role | Goals | Pain Points |
|---------|------|-------|-------------|
| **Nip (Hanif)** | ADMIN + Tech | Track 5 generasi, manage data lengkap | Data tersebar, manual tracking |
| **Abi/Mami** | ADMIN/EDITOR | Lihat cucu/cicit, tambah info keluarga | Gaptek, butuh UI simple |
| **Saudara (Poligami)** | EDITOR | Track anak dari istri berbeda | Kompleksitas relasi |
| **Keponakan** | VIEWER | Cari tahu hubungan keluarga | Tidak tahu silsilah lengkap |

### Age-Specific Personas

| Persona | Age | Tech Level | Role | Goals | Design Needs |
|---------|-----|------------|------|-------|--------------|
| **Remaja** | 14-20 | High | VIEWER | Cari tahu silsilah, share ke social media | Mobile-first, fast, aesthetic UI, touch gestures |
| **Young Adult** | 21-35 | Medium-High | EDITOR | Manage family data, export reports | Keyboard shortcuts, powerful search, bulk ops |
| **Middle Age** | 36-50 | Medium | EDITOR | Track children/grandchildren | Clear labels, step-by-step wizards, tooltips |
| **Eyang (Elder)** | 50+ | Low | VIEWER | Lihat cucu/cicit, cari kontak keluarga | Large fonts, high contrast, simple nav, print option |

### Age Distribution Target
| Age Group | Target % | Primary Device | Key Feature |
|-----------|----------|----------------|-------------|
| 14-20 (Teen) | 15% | Mobile | Tree visualization, share |
| 21-35 (Young Adult) | 35% | Desktop/Mobile | CRUD, export, search |
| 36-50 (Middle Age) | 35% | Desktop | Family management |
| 50+ (Senior) | 15% | Tablet/Desktop | View-only, print, large text |

---

## 3. User Stories & Acceptance Criteria

### 3.1 Authentication & Onboarding

**US-001: Register User**
- *Sebagai user baru, saya ingin register agar bisa akses aplikasi*
- **AC:**
  - Form: email, password, confirm password
  - Validasi: email unik, password min 8 chars
  - Auto-create role: VIEWER (default)
  - Email verifikasi via N8N + Brevo (optional untuk MVP)

**US-002: Login**
- *Sebagai user, saya ingin login agar bisa akses data keluarga*
- **AC:**
  - Input: email + password
  - JWT token expiry: 24 hours
  - Redirect ke dashboard setelah login

**US-003: Request ADMIN Access**
- *Sebagai user, saya ingin request akses ADMIN untuk manage keluarga*
- **AC:**
  - Form request dengan alasan
  - Notifikasi ke current ADMIN (via email/N8N)
  - ADMIN approve/reject dari dashboard

---

### 3.2 Family Management

**US-004: Create Family**
- *Sebagai ADMIN, saya ingin buat family baru agar bisa mulai tracking silsilah*
- **AC:**
  - Input: family name, head of family (link ke Person)
  - Auto-generate unique family ID
  - ADMIN menjadi owner family

**US-005: View Family Tree**
- *Sebagai user, saya ingin lihat tree silsilah 5 generasi*
- **AC:**
  - Visualisasi dari head of family → cicit (5 gen)
  - Zoom in/out, pan
  - Click person → detail panel
  - Load time <2s untuk 100 persons

**US-006: Export Family Data**
- *Sebagai user, saya ingin export data keluarga ke CSV/PDF*
- **AC:**
  - Format: List (bukan tree visual)
  - Columns: Name, gender, birthDate, deathDate, relationToHead, generation
  - Filter: by generation, by branch, all members
  - Download langsung

---

### 3.3 Person Management

**US-007: Add Person**
- *Sebagai EDITOR/ADMIN, saya ingin tambah anggota keluarga baru*
- **AC:**
  - Form: name, gender, birthDate, deathDate (optional), bio (optional)
  - Fuzzy duplicate detection: "Similar names exist: [list]"
  - Auto-link ke family (BELONGS_TO)
  - Skip foto upload untuk MVP

**US-008: Edit Person**
- *Sebagai EDITOR/ADMIN, saya ingin edit data anggota keluarga*
- **AC:**
  - Edit semua field kecuali id
  - Validasi: deathDate > birthDate
  - Audit log: updatedAt, updatedBy

**US-009: Delete Person**
- *Sebagai ADMIN, saya ingin hapus anggota keluarga (soft delete)*
- **AC:**
  - Soft delete (flag `isDeleted: true`)
  - Cascade: tidak hapus relasi terkait
  - Restore option (ADMIN only)

---

### 3.4 Relationship Management

**US-010: Add Marriage (Poligami Support)**
- *Sebagai EDITOR/ADMIN, saya ingin tambah relasi pernikahan (support poligami)*
- **AC:**
  - Select 2 persons (auto-complete dengan filter)
  - Input: startDate, isCurrent (default: true)
  - Auto-increment marriage order (1, 2, 3, 4)
  - Validasi: bukan ancestor/descendant/sibling
  - Support multiple current spouses (poligami)

**US-011: End Marriage (Divorce/Death)**
- *Sebagai ADMIN, saya ingin akhiri pernikahan (cerai/kematian)*
- **AC:**
  - Update: endDate, endReason ('divorce'|'death')
  - Set isCurrent: false
  - Auto-end jika spouse deathDate diisi (no remarriage)
  - Ex-spouse pindah ke detail panel (tidak tampil di tree)

**US-012: Add Parent-Child**
- *Sebagai EDITOR/ADMIN, saya ingin tambah relasi orang tua-anak*
- **AC:**
  - Select parent + child (auto-complete context-aware)
  - Validasi: child.birthDate > parent.birthDate + 10 years
  - Warning: child.birthDate > parent.deathDate + 9 months
  - Support single parent (1 parent only)

**US-013: Add Sibling**
- *Sebagai EDITOR/ADMIN, saya ingin tambah relasi saudara kandung*
- **AC:**
  - Auto-link via shared parents
  - Manual add: select 2+ persons with same parents
  - Validasi: bukan self, bukan ancestor/descendant

**US-014: Add Step-Parent**
- *Sebagai EDITOR/ADMIN, saya ingin tambah relasi orang tua tiri*
- **AC:**
  - Select step-parent + step-child
  - Validasi: step-parent married to biological parent
  - Relationship: `STEP_PARENT_OF` (dashed line di tree)
  - Tampil di tree dengan visual berbeda

---

### 3.5 Search & Discovery

**US-015: Search Person**
- *Sebagai user, saya ingin cari anggota keluarga by name*
- **AC:**
  - Input: min 3 chars
  - Real-time auto-complete (debounce 300ms)
  - Show: name, birthDate, parents (untuk disambiguate)
  - Filter: living/deceased toggle

**US-016: Find Relationship**
- *Sebagai user, saya ingin tahu hubungan antara 2 orang*
- **AC:**
  - Input: 2 persons
  - Output: path relasi (mis: "Sepupu 2 kali", "Paman")
  - Visual: highlight path di tree

**US-017: Smart Auto-Complete**
- *Sebagai user, saya ingin auto-complete yang context-aware saat tambah relasi*
- **AC:**
  - Filter candidate berdasarkan jenis relasi:
    - **Parent:** Exclude self, descendants, siblings
    - **Child:** Exclude self, ancestors, siblings
    - **Spouse:** Exclude self, ancestors, descendants, siblings
    - **Sibling:** Same generation (children of same parents)
  - Gender filter (untuk "Adek Perempuan")
  - Age ranking (kakak vs adek based on birthDate)

---

### 3.6 RBAC & Permissions

**US-018: Role-Based Access**
- *Sebagai system, saya ingin enforce RBAC untuk protect data*
- **AC:**
  - **ADMIN:** Full CRUD semua persons, manage users, transfer ownership
  - **EDITOR:** CRUD diri sendiri + descendants, view all
  - **VIEWER:** Read-only tree + search
  - Middleware check semua API endpoints

**US-019: Transfer Family Ownership**
- *Sebagai ADMIN, saya ingin transfer ownership jika head of family meninggal*
- **AC:**
  - Select new head of family (spouse/eldest child)
  - Update `headOfFamilyId`
  - Auto-transfer ADMIN role

---

## 4. Functional Requirements

### Must Have (MVP)
| ID | Feature | Priority |
|----|---------|----------|
| F-001 | Auth (register/login/JWT) | P0 |
| F-002 | Family CRUD | P0 |
| F-003 | Person CRUD (no foto) | P0 |
| F-004 | Marriage (poligami, divorce, death) | P0 |
| F-005 | Parent-Child relationship | P0 |
| F-006 | Sibling relationship | P0 |
| F-007 | Step-Parent relationship | P0 |
| F-008 | Tree visualization (5 gen) | P0 |
| F-009 | Search + auto-complete | P0 |
| F-010 | Smart auto-complete (context-aware) | P0 |
| F-011 | RBAC (ADMIN/EDITOR/VIEWER) | P0 |
| F-012 | Export CSV/PDF list | P1 |
| F-013 | Fuzzy duplicate detection | P1 |
| F-014 | Deceased tracking (deathDate) | P0 |
| F-015 | Date validation | P0 |

### Nice to Have (Phase 2)
| ID | Feature | Priority |
|----|---------|----------|
| F-016 | Foto upload (MinIO/S3) | P2 |
| F-017 | Email invite via N8N | P2 |
| F-018 | PWA (installable) | P2 |
| F-019 | Timeline view (events) | P3 |
| F-020 | Bulk import CSV | P3 |
| F-021 | Adoption tracking | P3 |
| F-022 | Mobile app (React Native) | P3 |

---

## 5. Non-Functional Requirements

### Performance
| Requirement | Target | Measurement |
|-------------|--------|-------------|
| Tree load time | <2s untuk 100 persons | Lighthouse |
| API response time | <500ms (95th percentile) | Prometheus |
| Auto-complete latency | <300ms (debounce) | Frontend metrics |
| Mobile load time | <2s (3G network) | Lighthouse Mobile |
| First Interactive | <1.5s | Lighthouse |
| Touch response | <50ms | Chrome DevTools |

### Security
| Requirement | Target | Implementation |
|-------------|--------|----------------|
| Password hashing | bcrypt (cost: 12) | Go bcrypt library |
| JWT expiry | 24 hours | JWT middleware |
| HTTPS | Required | Traefik SSL (Let's Encrypt) |
| Input validation | All endpoints | Go validator |
| SQL injection prevention | Parameterized queries | Neo4j driver |

### Availability & Scalability
| Requirement | Target | Notes |
|-------------|--------|-------|
| Uptime | 99% (homelab) | Grafana monitoring |
| Max persons | 500+ | Neo4j indexed queries |
| Max generations | 5 (hard limit) | Query truncation |
| Concurrent users | 20+ | Load testing required |
| Backup frequency | Daily auto-backup | Neo4j dump |
| Recovery time | <1 hour | Documented procedure |

### Mobile & Accessibility
| Requirement | Target | WCAG Level |
|-------------|--------|------------|
| Responsive design | Mobile-first | N/A |
| Font size | 16px base, scalable to 200% | AA |
| Color contrast | 4.5:1 minimum | AA |
| Keyboard navigation | Full support | AA |
| Screen reader | ARIA labels | AA |
| Focus indicators | Visible outlines | AA |
| Touch targets | Min 44x44px | AA |

### Age-Specific Requirements
| Age Group | Requirement | Implementation |
|-----------|-------------|----------------|
| **Teen (14-20)** | Mobile-first, fast load | PWA, lazy loading |
| **Young Adult (21-35)** | Keyboard shortcuts, power features | Hotkeys, advanced search |
| **Middle Age (36-50)** | Clear labels, wizards | Step-by-step forms |
| **Senior (50+)** | Large fonts, high contrast | Font toggle, contrast mode |

---

## 6. Technical Specifications

### 6.1 Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Frontend  │────▶│  Backend API │────▶│   Neo4j     │
│  (React)    │◀────│    (Go)      │◀────│  (Graph DB) │
│  Tree Viz   │     │  Auth + RBAC │     │  5 Generasi │
│  PWA        │     │  N8N Hooks   │     │  Backup     │
└─────────────┘     └──────────────┘     └─────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │   Brevo      │
                    │   (SMTP)     │
                    └──────────────┘
```

### 6.2 Neo4j Schema

#### Nodes
```cypher
(:Person {
  id: uuid,
  name: string,
  gender: 'male'|'female'|'other',
  birthDate: date,
  deathDate: date?,
  title: string?,                    // Haji, Hj, Dr, Ir, etc.
  bio: string?,
  address: {                         // Embedded address (MVP)
    street: string,
    neighborhood: string?,           // RT/RW
    city: string,
    province: string,
    postalCode: string?,
    country: string,
    coordinates: {lat: float, lng: float}?,
    isPrimary: bool,
    validFrom: date?,
    validUntil: date?
  }?,
  phoneNumbers: [                    // Array of phone numbers
    {
      number: string,
      type: 'whatsapp'|'mobile'|'home'|'work',
      label: string?,
      isPrimary: bool,
      isVerified: bool,
      visibility: 'public'|'family_only'|'admin_only'|'private',
      addedAt: datetime
    }
  ],
  photoUrl: url?,                    // Phase 2
  isDeleted: bool,
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
  personId: uuid?,                   // Link to Person (optional)
  lastLoginAt: datetime?,
  createdAt: datetime
})
```

#### Relationships
```cypher
(:Person)-[:MARRIED_TO {
  startDate: date,
  endDate: date?,
  endReason: 'divorce'|'death'?,
  isCurrent: bool,
  order: number
}]->(:Person)

(:Person)-[:PARENT_OF]->(:Person)
(:Person)-[:CHILD_OF]->(:Person)
(:Person)-[:SIBLING_OF]->(:Person)
(:Person)-[:STEP_PARENT_OF {sinceDate: date}]->(:Person)
(:Person)-[:BELONGS_TO]->(:Family)
(:User)-[:CAN_EDIT {role: string, grantedAt: datetime}]->(:Family)
```

### 6.3 API Endpoints

#### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/refresh` | Refresh token |
| POST | `/api/auth/logout` | Logout |

#### Family
| Method | Endpoint | Description | Role |
|--------|----------|-------------|------|
| POST | `/api/families` | Create family | ADMIN |
| GET | `/api/families/:id` | Get family details | All |
| GET | `/api/families/:id/tree` | Get family tree (5 gen) | All |
| PUT | `/api/families/:id` | Update family | ADMIN |
| POST | `/api/families/:id/transfer` | Transfer ownership | ADMIN |

#### Persons
| Method | Endpoint | Description | Role |
|--------|----------|-------------|------|
| POST | `/api/persons` | Add person | EDITOR+ |
| GET | `/api/persons/:id` | Get person details | All |
| PUT | `/api/persons/:id` | Update person | EDITOR+ |
| DELETE | `/api/persons/:id` | Soft delete person | ADMIN |
| POST | `/api/persons/:id/relate` | Add relationship | EDITOR+ |
| GET | `/api/persons/search` | Search persons | All |

#### Relationships
| Method | Endpoint | Description | Role |
|--------|----------|-------------|------|
| POST | `/api/relasi/marriage` | Add marriage | EDITOR+ |
| PUT | `/api/relasi/marriage/:id/end` | End marriage | ADMIN |
| POST | `/api/relasi/parent-child` | Add parent-child | EDITOR+ |
| POST | `/api/relasi/sibling` | Add sibling | EDITOR+ |
| POST | `/api/relasi/step-parent` | Add step-parent | EDITOR+ |
| GET | `/api/relasi/between` | Find relationship | All |

#### Export
| Method | Endpoint | Description | Role |
|--------|----------|-------------|------|
| GET | `/api/export/csv` | Export CSV | All |
| GET | `/api/export/pdf` | Export PDF | All |

### 6.4 Smart Auto-Complete Logic

#### Context-Aware Filtering

| Relationship Type | Exclude | Include | Additional Filter |
|-------------------|---------|---------|-------------------|
| **Parent** | Self, descendants, siblings | Non-descendants | Older generation preferred |
| **Child** | Self, ancestors, siblings | Non-ancestors | Younger generation preferred |
| **Spouse** | Self, ancestors, descendants, siblings | Non-blood relatives | Exclude current spouses |
| **Sibling** | Self, ancestors, descendants | Same parents | Same generation |
| **Step-Parent** | Self, ancestors, descendants | Married to biological parent | Current spouse only |

#### Cypher Query Example (Add Sister)
```cypher
MATCH (p:Person {id: $personId})
MATCH (candidate:Person)
WHERE candidate.id <> p.id
AND NOT (candidate)-[:PARENT_OF*]->(p)      // Bukan ancestor
AND NOT (p)-[:PARENT_OF*]->(candidate)      // Bukan descendant
AND NOT (candidate)-[:MARRIED_TO {isCurrent: true}]-(p)  // Bukan spouse
AND candidate.gender = 'female'             // Gender filter
AND EXISTS((candidate)<-[:PARENT_OF]-(parent)<-[:PARENT_OF]-(p))  // Share parent
RETURN candidate.name, candidate.birthDate
ORDER BY candidate.birthDate DESC           // Younger first (adek)
```

### 6.5 Age-Specific UI Guidelines

#### Teen (14-20 tahun)
| UI Element | Design Decision |
|------------|-----------------|
| **Navigation** | Hamburger menu, bottom nav on mobile |
| **Forms** | Inline validation, auto-save drafts |
| **Tree View** | Touch gestures (pinch zoom, swipe) |
| **Search** | Auto-complete with avatars |
| **Help** | Video tutorials (YouTube-style) |
| **Actions** | Quick actions (floating button) |
| **Feedback** | Toast notifications, animations |

#### Young Adult (21-35 tahun)
| UI Element | Design Decision |
|------------|-----------------|
| **Navigation** | Sidebar + breadcrumbs |
| **Forms** | Keyboard shortcuts, bulk edit |
| **Tree View** | Zoom controls, filter toggles |
| **Search** | Advanced filters, saved searches |
| **Help** | Documentation, keyboard shortcut list |
| **Actions** | Context menus, right-click actions |
| **Feedback** | Progress bars, batch operation status |

#### Middle Age (36-50 tahun)
| UI Element | Design Decision |
|------------|-----------------|
| **Navigation** | Large tabs, clear back button |
| **Forms** | Step-by-step wizard, progress indicator |
| **Tree View** | Click-to-expand, clear legends |
| **Search** | Simple search bar + filters |
| **Help** | Tooltips everywhere, FAQ section |
| **Actions** | Large buttons, clear labels |
| **Feedback** | Clear success/error messages |

#### Senior / Eyang (50+ tahun)
| UI Element | Design Decision |
|------------|-----------------|
| **Navigation** | Large tabs, persistent back button |
| **Forms** | One question per page, large inputs |
| **Tree View** | Click-to-expand, high contrast |
| **Search** | Large search bar, recent searches |
| **Help** | WhatsApp support link, phone number |
| **Actions** | Extra-large buttons, icons + text |
| **Feedback** | Large modals, print confirmation |
| **Special** | Font size toggle (A+ button), high contrast mode |

### 6.6 Accessibility Features

#### Must-Have (MVP)
- [ ] Font size toggle (100%, 125%, 150%, 200%)
- [ ] High contrast mode toggle
- [ ] Keyboard navigation (Tab, Enter, Esc)
- [ ] Focus indicators on all interactive elements
- [ ] ARIA labels on all buttons/icons
- [ ] Alt text on all images
- [ ] Form labels associated with inputs
- [ ] Error messages linked to form fields

#### Nice-to-Have (Phase 2)
- [ ] Screen reader optimization
- [ ] Voice control support
- [ ] Reduced motion mode
- [ ] Dyslexia-friendly font option
- [ ] Print stylesheet (senior-friendly layout)

### 6.7 Edge Cases Handling

| Edge Case | Validation | Error Message |
|-----------|------------|---------------|
| **Incest** | Check path between persons | "Tidak dapat menambah relasi: hubungan darah terdeteksi" |
| **Circular Parent** | Check `NOT (child)-[:PARENT_OF*]->(parent)` | "Loop terdeteksi: anak tidak bisa jadi orang tua ancestor" |
| **Self-Relationship** | Check `person1.id <> person2.id` | "Tidak dapat menambah relasi dengan diri sendiri" |
| **Duplicate Spouse** | Unique constraint on active marriage | "Orang ini sudah menjadi spouse aktif" |
| **Child Before Parent Birth** | Validate `child.birthDate > parent.birthDate + 10y` | "Tanggal lahir anak tidak valid (lebih tua dari orang tua)" |
| **Child After Parent Death** | Warning if `child.birthDate > parent.deathDate + 9m` | "Peringatan: anak lahir setelah orang tua meninggal (>9 bulan)" |
| **Death Before Birth** | Hard validate `deathDate > birthDate` | "Tanggal kematian tidak valid (sebelum lahir)" |

---

## 7. Success Metrics

### Core Metrics
| Metric | Target | Measurement | Frequency |
|--------|--------|-------------|-----------|
| **Data Completeness** | 5 generasi ter-track | Count persons per generation | Weekly |
| **Performance** | Tree load <2s | Lighthouse | Per deploy |
| **Data Integrity** | Zero circular parents | Automated test suite | Per commit |
| **User Adoption** | 10+ family members aktif dalam 1 bulan | User analytics | Monthly |
| **Backup** | Daily backup verified | Backup logs check | Daily |
| **Uptime** | 99% | Grafana monitoring | Weekly |

### Age-Group Adoption Targets
| Age Group | Target Users | Activation Rate | Retention (30-day) |
|-----------|--------------|-----------------|-------------------|
| **Teen (14-20)** | 3-5 users | 80% (login + view tree) | 60% (return weekly) |
| **Young Adult (21-35)** | 5-8 users | 90% (login + edit data) | 80% (active weekly) |
| **Middle Age (36-50)** | 5-8 users | 85% (login + manage family) | 75% (active weekly) |
| **Senior (50+)** | 2-4 users | 70% (login + view/print) | 50% (return monthly) |

### Accessibility Metrics
| Metric | Target | Tool |
|--------|--------|------|
| **WCAG Compliance** | Level AA | axe-core |
| **Font Scaling** | Works at 200% | Manual testing |
| **Keyboard Nav** | 100% navigable | Manual testing |
| **Color Contrast** | 4.5:1 minimum | Contrast checker |
| **Mobile Usability** | 90+ score | Lighthouse Mobile |

### Performance by Age Group
| Age Group | Primary Metric | Target |
|-----------|----------------|--------|
| **Teen** | Mobile load time | <2s (3G) |
| **Young Adult** | Task completion time | <5 mins (add person) |
| **Middle Age** | Error rate | <5% (form submission) |
| **Senior** | Help requests | <2 per user (first week) |

---

## 8. Timeline & Milestones

### Development Phases

| Phase | Duration | Deliverables | Success Criteria |
|-------|----------|--------------|------------------|
| **Phase 1: Foundation** | Week 1-2 | Neo4j Docker, Backend scaffold, Auth JWT | Auth flow working, Neo4j queries tested |
| **Phase 2: Core API** | Week 2-3 | Family/Person/Relasi endpoints, RBAC | All CRUD endpoints functional |
| **Phase 3: Frontend** | Week 3-5 | Tree viz (react-flow), Forms, Search | Tree render <2s, auto-complete working |
| **Phase 4: Polish** | Week 5-6 | Testing, Docker Compose, Deploy, Docs | 95% test coverage, deployed to homelab |

### Age-Specific Testing Schedule

| Phase | Age Group | Test Activities | Success Criteria |
|-------|-----------|-----------------|------------------|
| **Phase 3 (Week 4)** | Young Adult (21-35) | Usability testing (keyboard nav, search) | Task completion <5 mins |
| **Phase 3 (Week 4)** | Teen (14-20) | Mobile testing (touch gestures, load time) | Mobile Lighthouse 90+ |
| **Phase 4 (Week 5)** | Middle Age (36-50) | Form wizard testing, error handling | Error rate <5% |
| **Phase 4 (Week 5)** | Senior (50+) | Font scaling, contrast, print view | Help requests <2 per user |

### Accessibility Audit Timeline

| Week | Activity | Tool | Target |
|------|----------|------|--------|
| Week 3 | Automated scan (axe-core) | axe-core | 0 critical issues |
| Week 4 | Keyboard navigation test | Manual | 100% navigable |
| Week 5 | Screen reader test | NVDA/VoiceOver | All labels read |
| Week 5 | Color contrast check | Contrast checker | 4.5:1 minimum |
| Week 6 | Font scaling test | Browser zoom | Works at 200% |

### Deployment Milestones

| Milestone | Date | Deliverable | Stakeholders |
|-----------|------|-------------|--------------|
| **Alpha** | Week 4 | Core features (CRUD + tree) | Nip + 2 testers |
| **Beta** | Week 6 | All features + accessibility | Family members (10+) |
| **RC** | Week 7 | Bug fixes, performance | All age groups |
| **Launch** | Week 8 | Production deploy | All family members |

---

## 9. Risks & Mitigation

### Technical Risks

| Risk | Impact | Probability | Mitigation | Owner |
|------|--------|-------------|------------|-------|
| **Neo4j complexity** | High | Medium | Start dengan query simple, test di Docker dulu | Backend |
| **Tree viz performance** | Medium | Medium | Limit initial render (100 nodes), lazy loading | Frontend |
| **RBAC edge cases** | Medium | High | Write test cases early, manual testing | Backend |
| **Polygami UI complexity** | Medium | High | Show current spouses inline, ex-spouses in panel | Frontend |
| **Foto storage** | Low | Low | Skip for MVP, local FS nanti | Backend |
| **Email invite** | Low | Low | N8N integration Phase 2 | Backend |

### Age-Specific Risks

| Risk | Impact | Probability | Mitigation | Owner |
|------|--------|-------------|------------|-------|
| **Senior adoption low** | High | Medium | WhatsApp support, large fonts, print option | Frontend |
| **Teen boredom** | Medium | Medium | Fast load, mobile-first, aesthetic design | Frontend |
| **Accessibility non-compliance** | High | Low | Early audit (Week 3), fix before Beta | Both |
| **Mobile performance poor** | High | Medium | Lighthouse testing per commit, optimize images | Frontend |
| **Complex forms scare seniors** | Medium | High | Step-by-step wizard, progress indicators | Frontend |
| **Keyboard shortcuts conflict** | Low | Low | Document shortcuts, allow disable | Frontend |

### Mitigation Priority Matrix

| Priority | Risk | Action Required By |
|----------|------|-------------------|
| **P0** | Senior adoption low | Week 3 (Frontend design) |
| **P0** | Accessibility compliance | Week 3 (Automated audit) |
| **P1** | Mobile performance | Week 4 (Lighthouse testing) |
| **P1** | Tree viz performance | Week 4 (Lazy loading) |
| **P2** | Teen engagement | Week 5 (UI polish) |
| **P2** | Form complexity | Week 5 (Wizard implementation) |

---

## 10. Open Questions

### Technical Decisions
| Question | Decision | Notes |
|----------|----------|-------|
| **Foto storage** | Skip MVP | Local FS/MinIO Phase 2 |
| **Email invite** | N8N + Brevo | Phase 2 |
| **PWA** | Bonus | Mobile responsive first |
| **Adoption tracking** | Skip | Adoption = biological |
| **Step-parent** | ✅ Include MVP | `STEP_PARENT_OF` relationship |
| **Export format** | CSV/PDF list | Bukan tree visual |
| **Duplicate detection** | Fuzzy (Levenshtein) | Min 80% similarity threshold |

### Age-Specific Decisions
| Question | Decision | Notes |
|----------|----------|-------|
| **Font size toggle** | ✅ Include MVP | 100%, 125%, 150%, 200% |
| **High contrast mode** | ✅ Include MVP | Toggle button in header |
| **Print view** | ✅ Include MVP | Senior-friendly layout |
| **Video tutorials** | Phase 2 | YouTube-style for teens |
| **Keyboard shortcuts** | Phase 2 | Power users (young adults) |
| **WhatsApp support** | ✅ Include MVP | Floating button for seniors |
| **Touch gestures** | ✅ Include MVP | Pinch zoom, swipe on mobile |

### Pending Decisions
| Question | Options | Decision Needed By |
|----------|---------|-------------------|
| **Default font family** | Inter vs System UI | Phase 1 (Foundation) |
| **Color scheme** | Light/Dark toggle or Light only | Phase 3 (Frontend) |
| **Onboarding flow** | Interactive tour or Skip-able hints | Phase 3 (Frontend) |
| **Share feature** | Social media export or Internal only | Phase 2 |

---

## 11. Appendices

### A. Cypher Query Reference
Lihat: `docs/queries.md` (TODO)

### B. API Schema
Lihat: `docs/api.md` (TODO)

### C. Frontend Components
Lihat: `IDEATION.md` - Frontend Components section

### D. Cultural Notes
- **Poligami:** Support hingga 4 istri (Islamic context)
- **Adoption:** Diperlakukan sama dengan anak kandung (Islamic law)
- **Naming:** Support title (Haji, Hj, Dr) separate dari name
- **Family Name:** Optional (Indonesian context, no surname tradition)

### E. Age-Inclusive Design Notes
- **Teen (14-20):** Mobile-first, fast load, aesthetic UI, touch gestures
- **Young Adult (21-35):** Keyboard shortcuts, power features, export options
- **Middle Age (36-50):** Clear labels, step-by-step wizards, tooltips
- **Senior (50+):** Large fonts (toggle), high contrast, print option, WhatsApp support
- **Accessibility:** WCAG 2.1 Level AA compliance (font scaling, keyboard nav, ARIA labels)

---

## 11b. Changelog

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| **1.0** | March 21, 2026 | Initial PRD (core features, Neo4j schema, API design) | Hanif Naufal Ashari |
| **1.1** | March 21, 2026 | Age-inclusive UX update (14-50+ years target) | Hanif Naufal Ashari |

### v1.1 Changes (Age-Inclusive Update)
- ✅ Added age-specific personas (Remaja, Young Adult, Middle Age, Eyang)
- ✅ Added age distribution targets (15% teen, 35% young adult, 35% middle age, 15% senior)
- ✅ Updated Non-Functional Requirements with accessibility specs (WCAG 2.1 AA)
- ✅ Added Age-Specific UI Guidelines section (6.5)
- ✅ Added Accessibility Features section (6.6)
- ✅ Updated Success Metrics with age-group adoption targets
- ✅ Updated Timeline with age-specific testing schedule
- ✅ Updated Risks with age-specific risks + mitigation priority matrix
- ✅ Updated Open Questions with age-specific decisions
- ✅ Added Cultural + Age-Inclusive Design Notes to Appendices

---

## 12. Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| **Product Owner** | Hanif Naufal Ashari | TBD | - |
| **Tech Lead** | TBD | TBD | - |
| **QA Lead** | TBD | TBD | - |

---

**Next Steps:**
1. ✅ PRD Approved
2. ⏳ Setup Neo4j Docker container
3. ⏳ Backend scaffold (Go + Gin)
4. ⏳ Frontend prototype (React + react-flow)

---

*Last Updated: March 21, 2026*
