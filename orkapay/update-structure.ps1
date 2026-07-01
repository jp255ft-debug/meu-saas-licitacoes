# update-structure.ps1
# Gera/atualiza o arquivo estrutura_projeto.txt com a arvore de diretorios
# Uso: .\update-structure.ps1 [-MaxDepth 4] [-CacheMinutes 60]

param(
    [int]$MaxDepth = 4,
    [int]$CacheMinutes = 60
)

$ProjectRoot = "C:\Source\Repos\orkapay"
$Output = "$ProjectRoot\estrutura_projeto.txt"
$ExcludeDirs = @('node_modules', '.git', '__pycache__', 'dist', 'build', 'venv', '.next', '.vscode', 'bin', 'obj', '.terraform', 'vendor')

# Verificar cache
if (Test-Path $Output) {
    $LastWrite = (Get-Item $Output).LastWriteTime
    $AgeMinutes = ((Get-Date) - $LastWrite).TotalMinutes
    if ($AgeMinutes -lt $CacheMinutes) {
        Write-Host "Estrutura atualizada ha $([math]::Round($AgeMinutes))min. Pulando..." -ForegroundColor Yellow
        exit 0
    }
}

Write-Host "Atualizando estrutura do projeto..." -ForegroundColor Cyan

# Funcao para gerar a arvore recursivamente
function Get-Tree {
    param(
        [string]$Path,
        [int]$CurrentDepth = 0,
        [int]$MaxDepth
    )

    if ($CurrentDepth -ge $MaxDepth) { return }

    $items = Get-ChildItem -Path $Path -Directory | Where-Object {
        $_.Name -notin $ExcludeDirs
    } | Sort-Object Name

    $count = $items.Count
    $i = 0

    foreach ($item in $items) {
        $i++
        if ($i -eq $count) {
            $prefix = "+-- "
        } else {
            $prefix = "|-- "
        }
        
        if ($CurrentDepth -gt 0) {
            $indent = "    " * ($CurrentDepth - 1) + "|   "
        } else {
            $indent = ""
        }
        
        # Linha do diretorio
        $line = "$indent$prefix$($item.Name)"
        $line

        # Recursao nos subdiretorios
        Get-Tree -Path $item.FullName -CurrentDepth ($CurrentDepth + 1) -MaxDepth $MaxDepth
    }
}

# Montar o conteudo
$header = @"
# Estrutura do Projeto OrkaPay
# Gerado em: $(Get-Date -Format 'dd/MM/yyyy HH:mm:ss')
# Profundidade maxima: $MaxDepth niveis
# Ignora: $($ExcludeDirs -join ', ')
# ==========================================

"@

$tree = Get-Tree -Path $ProjectRoot -MaxDepth $MaxDepth
$content = $header + "orkapay/" + [Environment]::NewLine + ($tree -join [Environment]::NewLine)

# Salvar
$content | Out-File -FilePath $Output -Encoding UTF8

Write-Host "Estrutura atualizada em $Output" -ForegroundColor Green
Write-Host "Total de diretorios listados: $(($tree | Measure-Object | Select-Object -ExpandProperty Count))" -ForegroundColor Cyan
