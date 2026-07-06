#!/bin/bash
# onboard-client.sh - Onboarding manual de cliente OrkaPay
# Uso: bash scripts/onboard-client.sh
set -e

echo ""
echo "=========================================="
echo "   🚀 OrkaPay - Onboarding de Cliente"
echo "=========================================="
echo ""

# 1. Coletar dados
read -p "Nome da Empresa: " COMPANY_NAME
read -p "CNPJ (apenas números): " CNPJ
read -p "E-mail do contato: " EMAIL
read -p "Telefone (opcional): " PHONE

CLIENT_ID=$(echo "$COMPANY_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')
API_KEY=$(openssl rand -base64 32 | tr -d '/+=' | head -c 32)

echo ""
echo "📋 Dados do Cliente:"
echo "  Empresa:  $COMPANY_NAME"
echo "  ClientID: $CLIENT_ID"
echo "  CNPJ:     $CNPJ"
echo "  E-mail:   $EMAIL"
echo ""

# 2. Validar CNPJ (básico)
if [ ${#CNPJ} -eq 14 ]; then
    echo "✅ CNPJ válido (14 dígitos)"
else
    echo "⚠️  CNPJ com formato inesperado (${#CNPJ} dígitos). Verifique."
fi

# 3. Gerar chaves
echo ""
echo "🔑 Gerando API Key..."
echo "  API Key: $API_KEY"

# 4. Instruções para inserir no banco
echo ""
echo "📝 Para ativar, execute no banco:"
echo ""
echo "INSERT INTO api_keys (id, client_id, client_name, key_hash, created_at)"
echo "VALUES (gen_random_uuid(), '$CLIENT_ID', '$COMPANY_NAME', crypt('$API_KEY', gen_salt('bf')), NOW());"
echo ""

# 5. Template de e-mail
echo "=========================================="
echo "   📧  E-mail a ser enviado"
echo "=========================================="
echo ""
echo "Assunto: Bem-vindo ao OrkaPay - Suas Credenciais"
echo ""
echo "Olá $COMPANY_NAME,"
echo ""
echo "Seu cadastro foi aprovado! Aqui estão suas credenciais:"
echo ""
echo "🔑 API Key: $API_KEY"
echo "🌐 Base URL: https://api.orkapay.com"
echo "📚 Docs: https://docs.orkapay.com"
echo ""
echo "Para testar:"
echo "curl -H 'X-API-Key: $API_KEY' https://api.orkapay.com/health"
echo ""
echo "Equipe OrkaPay"
echo "support@orkapay.com"
echo ""
echo "=========================================="
echo "✅ Onboarding concluído!"
echo "=========================================="