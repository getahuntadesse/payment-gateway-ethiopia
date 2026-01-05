# Ethiopian Payment Gateway Test Script
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  Ethiopian Payment Gateway Test Suite    " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "1. Testing Health Check..." -ForegroundColor Yellow
$health = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method Get
Write-Host "   Status: $($health.status)" -ForegroundColor Green
Write-Host "   Ethiopian Time: $($health.ethiopian_time)" -ForegroundColor Green
Write-Host ""

# Test 2: List Ethiopian Banks
Write-Host "2. Listing Ethiopian Banks..." -ForegroundColor Yellow
$banks = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/banks" -Method Get
foreach ($bank in $banks.banks) {
    Write-Host "   $($bank.code): $($bank.name)" -ForegroundColor White
}
Write-Host ""

# Test 3: Create Sample Ethiopian Payments
Write-Host "3. Creating Ethiopian Payments..." -ForegroundColor Yellow

# Payment 1: ETB Payment to CBE
$payment1 = @{
    amount = 5000.50
    currency = "ETB"
    reference = "CBE-" + (Get-Date -Format "yyyyMMddHHmmss")
    description = "የቡና ምርት ክፍያ (Coffee product payment)"
    customer_name = "የሱፍ አህመድ"
    bank_code = "CBE"
} | ConvertTo-Json

Write-Host "   Creating CBE Payment..." -ForegroundColor Gray
$response1 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/payments" -Method Post -Body $payment1 -ContentType "application/json"
Write-Host "   Payment ID: $($response1.payment_id)" -ForegroundColor Green
Write-Host "   Status: $($response1.status)" -ForegroundColor Green
Write-Host "   Message: $($response1.message)" -ForegroundColor Green
Write-Host ""

# Payment 2: USD Payment to Awash Bank
$payment2 = @{
    amount = 250.75
    currency = "USD"
    reference = "AWASH-" + (Get-Date -Format "yyyyMMddHHmmss")
    description = "Software license renewal"
    customer_name = "Tech Solutions PLC"
    bank_code = "AWASH"
} | ConvertTo-Json

Write-Host "   Creating Awash Bank Payment..." -ForegroundColor Gray
$response2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/payments" -Method Post -Body $payment2 -ContentType "application/json"
Write-Host "   Payment ID: $($response2.payment_id)" -ForegroundColor Green
Write-Host "   Status: $($response2.status)" -ForegroundColor Green
Write-Host ""

# Test 4: List Payments
Write-Host "4. Listing All Payments..." -ForegroundColor Yellow
$payments = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/payments" -Method Get
Write-Host "   Total Payments: $($payments.total)" -ForegroundColor Green
foreach ($payment in $payments.payments) {
    Write-Host "   - $($payment.reference): $($payment.amount) $($payment.currency) [$($payment.status)]" -ForegroundColor White
}
Write-Host ""

# Test 5: Get Statistics
Write-Host "5. Getting Statistics..." -ForegroundColor Yellow
$stats = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/statistics" -Method Get
Write-Host "   Total ETB: $($stats.total_amount_etb) Br" -ForegroundColor Green
Write-Host "   Total USD: $($stats.total_amount_usd) $" -ForegroundColor Green
Write-Host "   Successful: $($stats.successful_payments)" -ForegroundColor Green
Write-Host "   Pending: $($stats.pending_payments)" -ForegroundColor Yellow
Write-Host "   Failed: $($stats.failed_payments)" -ForegroundColor Red
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  Test Suite Completed Successfully!      " -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan