# Mapa de Estrutura do Projeto

## Regra Obrigatória
Antes de criar, mover ou deletar pastas/arquivos:
1. Execute `.\update-structure.ps1` (PowerShell)
2. Leia o arquivo `estrutura_projeto.txt` atualizado
3. Use o mapa como referência para localização de arquivos

## Arquivo de estrutura
- `estrutura_projeto.txt` — árvore de diretórios (máx. 4 níveis)
- Ignora: node_modules, .git, dist, build, vendor, bin, obj, .terraform
- Cache: 60 minutos (não regera se atualizado há < 1h)

## Manutenção
- O script `update-structure.ps1` na raiz do projeto faz a atualização
- Se o script não existir, crie-o antes de prosseguir
- Para forçar atualização: `.\update-structure.ps1 -CacheMinutes 0`

## Quando usar
- **SEMPRE** antes de criar novas pastas/arquivos
- **SEMPRE** antes de mover ou renomear diretórios
- **SEMPRE** antes de deletar arquivos/pastas
- **NUNCA** pule esta etapa — o mapa é a fonte da verdade para localização
