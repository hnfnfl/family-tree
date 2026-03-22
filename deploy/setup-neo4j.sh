#!/bin/bash
# Neo4j Setup Script for Keluarga Tree
# Run on lab2: bash ~/keluarga-tree/deploy/setup-neo4j.sh

NEO4J_URL="http://localhost:7474/db/neo4j/tx/commit"
NEO4J_USER="neo4j"
NEO4J_PASS="KeluargaTree2026!"

run_query() {
    local query="$1"
    curl -s -u "${NEO4J_USER}:${NEO4J_PASS}" \
        -X POST "${NEO4J_URL}" \
        -H "Content-Type: application/json" \
        -d "{\"statements\": [{\"statement\": \"${query}\"}]}"
}

echo "🔧 Setting up Neo4j for Keluarga Tree..."

# Create constraints
echo "📋 Creating constraints..."
run_query "CREATE CONSTRAINT person_id IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE"
run_query "CREATE CONSTRAINT family_id IF NOT EXISTS FOR (f:Family) REQUIRE f.id IS UNIQUE"
run_query "CREATE CONSTRAINT user_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE"

# Create indexes
echo "📊 Creating indexes..."
run_query "CREATE INDEX person_name IF NOT EXISTS FOR (p:Person) ON (p.name)"
run_query "CREATE INDEX person_birth_date IF NOT EXISTS FOR (p:Person) ON (p.birthDate)"

# Insert test data
echo "🧪 Inserting test data..."
run_query "
CREATE (head:Person {
  id: randomUUID(),
  name: 'Ahmad Ashari',
  gender: 'male',
  birthDate: date('1960-01-01'),
  bio: 'Kepala Keluarga',
  isDeleted: false,
  createdAt: datetime(),
  updatedAt: datetime()
})
RETURN head.id as headId, head.name as headName
"

echo "✅ Neo4j setup complete!"
echo "🌐 Browser: http://localhost:7474"
echo "🔌 Bolt: bolt://localhost:7687"
