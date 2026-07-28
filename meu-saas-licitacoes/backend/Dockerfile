FROM python:3.11-slim

WORKDIR /app

# Instala dependências do sistema (psycopg2)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq-dev gcc \
    && rm -rf /var/lib/apt/lists/*

# Copia e instala dependências Python
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copia o código do app
COPY . .

# Expõe a porta da aplicação
EXPOSE 8000

# Comando para iniciar o servidor
CMD ["uvicorn", "backend.app.main:app", "--host", "0.0.0.0", "--port", "8000"]