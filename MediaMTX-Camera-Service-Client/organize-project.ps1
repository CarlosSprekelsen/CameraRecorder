# organize-project.ps1
# This script sets up the directory structure and moves existing documentation stubs into place.

# Define directories to create
$dirs = @(
    "docs\architecture",
    "docs\api",
    "docs\development",
    "public",
    "src\components",
    "src\hooks",
    "src\services",
    "src\stores",
    "src\utils",
    "tests\unit",
    "tests\integration",
    "tests\e2e"
)

# Create directories if they don't exist
foreach ($dir in $dirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
}

# Map source files to their target locations
$stubMoves = @{  
    "client-architecture.md"             = "docs\architecture\client-architecture.md"
    "client-api-reference.md"            = "docs\api\client-api-reference.md"
    "client-coding-standards.md"         = "docs\development\client-coding-standards.md"
    "client-documentation-guidelines.md" = "docs\development\client-documentation-guidelines.md"
    "testing-guidelines.md"              = "docs\development\testing-guidelines.md"
    "CONTRIBUTING.md"                    = "docs\development\contributing.md"
    "ci-cd.md"                           = "docs\development\ci-cd.md"
    "deployment.md"                      = "docs\deployment.md"
}

# Move each stub, warning if missing
foreach ($source in $stubMoves.Keys) {
    $destination = $stubMoves[$source]
    if (Test-Path $source) {
        Move-Item -Path $source -Destination $destination -Force
    } else {
        Write-Host "Warning: '$source' not found in project root." -ForegroundColor Yellow
    }
}

Write-Host "Project structure organized successfully." -ForegroundColor Green
