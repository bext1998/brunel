$ErrorActionPreference = 'Stop'

$credentialEnvNames = @(
    'ANTHROPIC_API_KEY',
    'OPENAI_API_KEY',
    'OPENROUTER_API_KEY',
    'GEMINI_API_KEY'
)

foreach ($name in $credentialEnvNames) {
    $value = [Environment]::GetEnvironmentVariable($name)
    $state = if ([string]::IsNullOrWhiteSpace($value)) { 'absent' } else { 'present' }
    Write-Output "environment.$name=$state"
}

$settingsPath = Join-Path $env:USERPROFILE '.pi\agent\settings.json'
if (-not (Test-Path -LiteralPath $settingsPath -PathType Leaf)) {
    Write-Output 'pi_settings=absent'
} else {
    try {
        $settings = Get-Content -LiteralPath $settingsPath -Raw | ConvertFrom-Json
        $providerNames = @()
        if ($null -ne $settings.provider) {
            $providerNames += [string]$settings.provider
        }
        if ($null -ne $settings.providers) {
            $providerNames += @($settings.providers.psobject.Properties.Name)
        }
        $providerNames = @($providerNames | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | Sort-Object -Unique)
        if ($providerNames.Count -eq 0) {
            Write-Output 'pi_settings=present;providers=none-detected'
        } else {
            Write-Output ("pi_settings=present;providers=" + ($providerNames -join ','))
        }
    } catch {
        Write-Output 'pi_settings=present;parse=failed'
    }
}

$credentialManagerEntries = & cmdkey.exe /list 2>$null
$cmdkeyExitCode = $LASTEXITCODE
if ($cmdkeyExitCode -ne 0) {
    Write-Output "credential_manager=query-failed;exit=$cmdkeyExitCode"
} else {
    $matched = @($credentialManagerEntries | Where-Object {
        $_ -match '(?i)(anthropic|openai|openrouter|gemini|google|pi-coding)'
    })
    $state = if ($matched.Count -eq 0) { 'no-matching-entry' } else { 'matching-entry-present' }
    Write-Output "credential_manager=$state"
}
