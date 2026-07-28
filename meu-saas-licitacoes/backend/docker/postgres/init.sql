-- Script de inicialização do PostgreSQL
-- Cria extensão UUID e schema base

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Nota: As tabelas são criadas pelo SQLAlchemy via Alembic.
-- Este script é usado apenas para configurações iniciais do banco.