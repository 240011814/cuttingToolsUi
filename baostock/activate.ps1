# activate.ps1 - 启动 baostock Python 虚拟环境

$venvPath = Join-Path $PSScriptRoot ".venv"

if (-not (Test-Path $venvPath)) {
    Write-Host "Creating virtual environment..." -ForegroundColor Cyan
    python -m venv $venvPath
}

& "$venvPath\Scripts\Activate.ps1"

Write-Host "Installing dependencies..." -ForegroundColor Cyan
pip install -r (Join-Path $PSScriptRoot "requirements.txt") -i https://pypi.org/simple

Write-Host "Virtual environment ready!" -ForegroundColor Green