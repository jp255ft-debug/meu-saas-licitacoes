// Script de inicialização do MongoDB
// Cria o usuário admin e o banco licitacao_db

db = db.getSiblingDB("licitacao_db");

db.createUser({
  user: "admin",
  pwd: "admin_pass",
  roles: [
    { role: "readWrite", db: "licitacao_db" },
  ],
});

// Cria coleções com validação inicial
db.createCollection("editais_pncp", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["pncp_id", "orgao_nome", "modalidade", "objeto"],
      properties: {
        pncp_id: { bsonType: "string" },
        orgao_nome: { bsonType: "string" },
        modalidade: { bsonType: "string" },
        objeto: { bsonType: "string" },
      },
    },
  },
});

db.createCollection("processamento_ia", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["edital_mongodb_id", "resumo"],
      properties: {
        edital_mongodb_id: { bsonType: "string" },
        resumo: { bsonType: "string" },
      },
    },
  },
});

// Índices
db.editais_pncp.createIndex({ pncp_id: 1 }, { unique: true });
db.editais_pncp.createIndex({ data_publicacao: -1 });
db.editais_pncp.createIndex({ uf: 1, data_publicacao: -1 });
db.editais_pncp.createIndex({ modalidade: 1 });
db.editais_pncp.createIndex({ objeto: "text" });

db.processamento_ia.createIndex({ edital_mongodb_id: 1 }, { unique: true });

print("MongoDB inicializado com sucesso!");