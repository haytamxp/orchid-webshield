$ErrorActionPreference = "Stop"

$ExpectedToolchain = "go1.26.9"
$OriginalGoToolchain = $env:GOTOOLCHAIN
$Failures = New-Object System.Collections.Generic.List[string]

function Invoke-NativeCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,

        [Parameter(Mandatory = $true)]
        [string[]]$Arguments
    )

    # Windows PowerShell 5.1 can treat redirected native stderr as
    # PowerShell errors. Continue locally and decide using the exit code.
    $PreviousPreference = $ErrorActionPreference

    try {
        $ErrorActionPreference = "Continue"
        $Output = @(& $FilePath @Arguments 2>&1)
        $ExitCode = $LASTEXITCODE
    }
    catch {
        $Output = @($_.Exception.Message)
        $ExitCode = 127
    }
    finally {
        $ErrorActionPreference = $PreviousPreference
    }

    $Lines = @(
        $Output | ForEach-Object { [string]$_ }
    )

    return [PSCustomObject]@{
        Output   = $Lines
        ExitCode = $ExitCode
    }
}

function Invoke-GoCheck {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,

        [Parameter(Mandatory = $true)]
        [string[]]$GoArgs
    )

    Write-Host "`n--- $Name ---" -ForegroundColor Yellow

    $Result = Invoke-NativeCommand -FilePath "go" -Arguments $GoArgs

    foreach ($Line in $Result.Output) {
        Write-Host $Line
    }

    if ($Result.ExitCode -ne 0) {
        $Failures.Add("$Name failed with exit code $($Result.ExitCode).")
        Write-Host "FAIL: $Name" -ForegroundColor Red
    }
    else {
        Write-Host "PASS: $Name" -ForegroundColor Green
    }
}

try {
    $Repo = (Resolve-Path (Join-Path $PSScriptRoot "..\..\")).Path
    $BackendPath = Join-Path $Repo "backend"

    if (-not (Test-Path -LiteralPath (Join-Path $BackendPath "go.mod") -PathType Leaf)) {
        throw "Backend go.mod not found: $BackendPath"
    }

    # Pin this process to the Go version used by CI. Do not change the
    # persistent or system-wide Go configuration.
    $env:GOTOOLCHAIN = $ExpectedToolchain

    Write-Host "`n========== WEBSHIELD P4 LOCAL VALIDATION ==========" -ForegroundColor Cyan
    Write-Host "Repository: $Repo"
    Write-Host "Expected toolchain: $ExpectedToolchain"

    $VersionResult = Invoke-NativeCommand -FilePath "go" -Arguments @("version")

    foreach ($Line in $VersionResult.Output) {
        Write-Host $Line
    }

    if ($VersionResult.ExitCode -ne 0) {
        throw "Go toolchain selection failed with exit code $($VersionResult.ExitCode). Review the output above."
    }

    $VersionText = $VersionResult.Output -join "`n"

    if ($VersionText -notmatch '(?m)^go version go1\.26\.9(?:\s|$)') {
        throw "Expected Go 1.26.9, but the reported version did not match. Review the output above."
    }

    $RootResult = Invoke-NativeCommand -FilePath "go" -Arguments @("env", "GOROOT")

    foreach ($Line in $RootResult.Output) {
        Write-Host $Line
    }

    if ($RootResult.ExitCode -ne 0) {
        throw "Unable to determine the selected toolchain's GOROOT."
    }

    $GoRoot = ($RootResult.Output -join "`n").Trim()
    $GofmtPath = Join-Path $GoRoot "bin\gofmt.exe"

    if (-not (Test-Path -LiteralPath $GofmtPath -PathType Leaf)) {
        throw "gofmt was not found at $GofmtPath"
    }

    Push-Location -LiteralPath $BackendPath

    try {
        $GoFiles = @(
            Get-ChildItem -Path . -Recurse -File -Filter "*.go" |
                Where-Object {
                    $_.FullName -notmatch '[\\/]vendor[\\/]'
                }
        )

        if ($GoFiles.Count -eq 0) {
            throw "No backend Go source files found. Validation cannot establish code quality without source files."
        }

        Write-Host "`nGo source files found: $($GoFiles.Count)" -ForegroundColor Cyan

        Write-Host "`n--- gofmt (line-ending normalized) ---" -ForegroundColor Yellow

        $FormatTempRoot = Join-Path $env:TEMP ("webshield-gofmt-" + [guid]::NewGuid().ToString("N"))
        $NormalizedFiles = New-Object System.Collections.Generic.List[string]
        $SourceByTempPath = @{}
        $FormatUtf8 = New-Object System.Text.UTF8Encoding($false)

        try {
            New-Item -ItemType Directory -Path $FormatTempRoot | Out-Null

            $BackendPrefix = [System.IO.Path]::GetFullPath($BackendPath).TrimEnd('\') + '\'

            foreach ($GoFile in $GoFiles) {
                $SourcePath = [System.IO.Path]::GetFullPath($GoFile.FullName)

                if (-not $SourcePath.StartsWith(
                    $BackendPrefix,
                    [StringComparison]::OrdinalIgnoreCase
                )) {
                    throw "Source file is outside the backend directory: $SourcePath"
                }

                $RelativeGoPath = $SourcePath.Substring($BackendPrefix.Length)
                $TempPath = Join-Path $FormatTempRoot $RelativeGoPath
                $TempParent = Split-Path -Parent $TempPath

                if (-not (Test-Path -LiteralPath $TempParent -PathType Container)) {
                    New-Item -ItemType Directory -Path $TempParent -Force | Out-Null
                }

                $Text = [System.IO.File]::ReadAllText($SourcePath)
                $Text = [regex]::Replace($Text, "\r\n?", "`n")

                [System.IO.File]::WriteAllText(
                    $TempPath,
                    $Text,
                    $FormatUtf8
                )

                $NormalizedFiles.Add($TempPath)
                $SourceByTempPath[$TempPath] = $SourcePath
            }

            $FormatArgs = @("-l") + $NormalizedFiles.ToArray()
            $FormatResult = Invoke-NativeCommand -FilePath $GofmtPath -Arguments $FormatArgs

            if ($FormatResult.ExitCode -ne 0) {
                $Failures.Add("gofmt failed with exit code $($FormatResult.ExitCode).")
                $FormatResult.Output | ForEach-Object { Write-Host $_ }
            }
            else {
                $ReportedPaths = @(
                    $FormatResult.Output |
                        Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
                )

                if ($ReportedPaths.Count -gt 0) {
                    foreach ($ReportedPath in $ReportedPaths) {
                        $ReportedPath = $ReportedPath.Trim()

                        if ($SourceByTempPath.ContainsKey($ReportedPath)) {
                            Write-Host "Formatting required: $($SourceByTempPath[$ReportedPath])" -ForegroundColor Red
                        }
                        else {
                            Write-Host "Formatter output: $ReportedPath" -ForegroundColor Red
                        }
                    }

                    $Failures.Add("Go files require formatting after line-ending normalization.")
                }
                else {
                    Write-Host "PASS: gofmt (normalized temporary copies)" -ForegroundColor Green
                }
            }
        }
        finally {
            if (Test-Path -LiteralPath $FormatTempRoot) {
                Remove-Item -LiteralPath $FormatTempRoot -Recurse -Force -ErrorAction SilentlyContinue
            }
        }
        Invoke-GoCheck -Name "go vet" -GoArgs @(
            "vet", "-mod=readonly", "./..."
        )

        Invoke-GoCheck -Name "go test" -GoArgs @(
            "test", "-count=1", "-mod=readonly", "./..."
        )

        Invoke-GoCheck -Name "go build" -GoArgs @(
            "build", "-mod=readonly", "./..."
        )
    }
    finally {
        Pop-Location
    }

    Write-Host "`n========== VALIDATION SUMMARY ==========" -ForegroundColor Cyan

    if ($Failures.Count -gt 0) {
        $Failures | ForEach-Object {
            Write-Host "FAIL: $_" -ForegroundColor Red
        }

        throw "P4 local validation failed. Review the individual results."
    }

    Write-Host "PASS: local backend quality checks completed." -ForegroundColor Green
    Write-Host "Race detection and security scanners still require separate verification."
}
finally {
    if ($null -eq $OriginalGoToolchain) {
        Remove-Item Env:GOTOOLCHAIN -ErrorAction SilentlyContinue
    }
    else {
        $env:GOTOOLCHAIN = $OriginalGoToolchain
    }
}
