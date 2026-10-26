// Jalankan: mongosh portofolio.js
db = db.getSiblingDB("praktikum_backend");

// db.student_portfolios.drop();

db.createCollection("student_portfolios", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["student_id", "headline", "projects", "certificates", "created_at", "updated_at"],
      properties: {
        student_id: { bsonType: "number", minimum: 1, description: "merujuk students.id di PostgreSQL" },
        headline:   { bsonType: "string", minLength: 5, maxLength: 100 },
        bio:        { bsonType: "string", maxLength: 1000 },
        projects: {
          bsonType: "array",
          maxItems: 20,
          items: {
            bsonType: "object",
            required: ["_id", "title", "technologies", "year"],
            properties: {
              _id:          { bsonType: "objectId" },
              title:        { bsonType: "string", minLength: 3, maxLength: 100 },
              technologies: { bsonType: "array", maxItems: 15, items: { bsonType: "string" } },
              url:          { bsonType: "string" },
              year:         { bsonType: "number", minimum: 2000, maximum: 2100 }
            }
          }
        },
        certificates: {
          bsonType: "array",
          maxItems: 30,
          items: {
            bsonType: "object",
            required: ["name", "issuer", "issued_at"],
            properties: {
              name:      { bsonType: "string", minLength: 3, maxLength: 100 },
              issuer:    { bsonType: "string", minLength: 2, maxLength: 100 },
              issued_at: { bsonType: "date" }
            }
          }
        },
        created_at: { bsonType: "date" },
        updated_at: { bsonType: "date" }
      }
    }
  },
  validationLevel: "strict",
  validationAction: "error"
});

db.student_portfolios.createIndex(
  { student_id: 1 },
  { unique: true, name: "student_unique" }
);

db.student_portfolios.createIndex(
  { "projects.technologies": 1 },
  { name: "projects_technologies_idx" }
);

// Dokumen contoh (ganti student_id sesuai id mahasiswa yang ada di PostgreSQL)
db.student_portfolios.insertOne({
  student_id: 1,
  headline: "Backend developer Go",
  bio: "Mahasiswa D4 Teknik Informatika, fokus pada backend dan data.",
  projects: [
    {
      _id: new ObjectId(),
      title: "API Kampus",
      technologies: ["go", "postgresql", "mongodb"],
      url: "https://github.com/fahmi-blip/api-kampus",
      year: 2026
    }
  ],
  certificates: [
    { name: "Belajar Dasar Pemrograman Go", issuer: "Dicoding", issued_at: ISODate("2026-05-01T00:00:00Z") }
  ],
  created_at: new Date(),
  updated_at: new Date()
});

print("== index ==");
printjson(db.student_portfolios.getIndexes());
print("== dokumen ==");
printjson(db.student_portfolios.findOne({ student_id: 1 }));