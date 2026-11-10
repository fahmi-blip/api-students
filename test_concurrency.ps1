Add-Type -AssemblyName System.Net.Http

$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluMSIsInJvbGUiOiJ1c2VyIiwiaXNzIjoiYmFja2VuZCIsInN1YiI6IjkiLCJleHAiOjE3OTMwMTI1MjUsImlhdCI6MTc5MzAxMTYyNX0.C06ZFdAFaDcHtEUy6VC4iWoE3kgwdiKKyHpCoVH-N80"
$url = "http://localhost:3000/api/v1/students/1/portfolio/projects"

$client = [System.Net.Http.HttpClient]::new()
$client.DefaultRequestHeaders.Authorization = [System.Net.Http.Headers.AuthenticationHeaderValue]::new("Bearer", $token)

Write-Host "Mengirim 10 request secara serentak..." -ForegroundColor Cyan

$tasks = 1..10 | ForEach-Object {
    $content = [System.Net.Http.StringContent]::new(
        '{"title":"Proyek Uji Serentak","technologies":["Go","MongoDB"],"year":2026}',
        [System.Text.Encoding]::UTF8,
        "application/json"
    )
    $client.PostAsync($url, $content)
}

[System.Threading.Tasks.Task]::WaitAll($tasks)

Write-Host "`nHasil Status Code:" -ForegroundColor Green
$tasks | ForEach-Object {
    Write-Output "Status: $($_.Result.StatusCode.value__)"
}

$client.Dispose()
