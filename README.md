 Create Database
cmd
# Open Command Prompt as Administrator
cd C:\Program Files\PostgreSQL\15\bin

# Create database
psql -U postgres -c "CREATE DATABASE ethiopian_payments;"

# Verify database
psql -U postgres -c "\l"
 Run Migrations
cmd
# In your project directory
psql -U postgres -d ethiopian_payments -f migrations\001_init_schema.up.sql
Verify Database
cmd
# Connect to database
psql -U postgres -d ethiopian_payments

# Check tables
\dt

# Check sample data
SELECT * FROM payments LIMIT 5;

# Check statistics view
SELECT * FROM payment_statistics;
RabbitMQ Setup 🐇
Verify RabbitMQ is Running
cmd
# Open PowerShell as Administrator
Get-Service RabbitMQ

# Should show: Running
Access Management UI
Open browser: http://localhost:15672

Login: guest / guest

Check Queues tab - you should see ethiopian_payment_queue

Running the Application 
 First Terminal - Run API
cmd
# Navigate to project directory
cd C:\Users\[YourUsername]\payment-gateway-ethiopia

# Install dependencies
go mod download

# Run API server
go run cmd/api/main.go
Expected Output:

text
{"level":"info","msg":"Starting Ethiopian Payment Gateway API...","time":"2024-01-15 10:30:00 EAT"}
{"level":"info","msg":"የኢትዮጵያ ክፍያ ግብይት መተግበሪያ እየተጀመረ ነው...","time":"2024-01-15 10:30:00 EAT"}
{"level":"info","msg":"Connected to PostgreSQL database successfully","time":"2024-01-15 10:30:01 EAT"}
{"level":"info","msg":"Connected to RabbitMQ successfully","time":"2024-01-15 10:30:01 EAT"}
{"level":"info","msg":"Ethiopian Payment Gateway API is running","port":8080,"environment":"development","ethiopian_time":"13:30:01","time":"2024-01-15 10:30:01 EAT"}
7.2 Second Terminal - Run Worker
cmd
# Open new Command Prompt
cd C:\Users\[YourUsername]\payment-gateway-ethiopia

# Run worker
go run cmd/worker/main.go
Expected Output:

text
{"level":"info","msg":"Starting Ethiopian Payment Processor Worker...","time":"2024-01-15 10:30:05 EAT"}
{"level":"info","msg":"የኢትዮጵያ ክፍያ ሂደት ሠራተኛ እየተጀመረ ነው...","time":"2024-01-15 10:30:05 EAT"}
{"level":"info","msg":"Connected to PostgreSQL database","time":"2024-01-15 10:30:05 EAT"}
{"level":"info","msg":"Connected to RabbitMQ","time":"2024-01-15 10:30:05 EAT"}
{"level":"info","msg":"Ethiopian Payment Processor is running","workers":5,"queue":"ethiopian_payment_queue","max_retries":3,"ethiopian_time":"13:30:05","time":"2024-01-15 10:30:05 EAT"}
Testing with Ethiopian Data 🇪🇹
Test Scripts
Create scripts/test-ethiopian.ps1 (PowerShell script):

powershell
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
8.2 Manual Testing with CURL
Test 1: Health Check

cmd
curl http://localhost:8080/api/v1/health
Test 2: Create Ethiopian Payment

cmd
curl -X POST http://localhost:8080/api/v1/payments ^
  -H "Content-Type: application/json" ^
  -d "{\"amount\": 7500.25, \"currency\": \"ETB\", \"reference\": \"CBE-TEST-001\", \"description\": \"የቤት ኪራይ ክፍያ\", \"customer_name\": \"ማርቆስ ገብረመድህን\", \"bank_code\": \"CBE\"}"
Test 3: Get Payment by ID

cmd
# Replace {payment_id} with actual ID from response
curl http://localhost:8080/api/v1/payments/{payment_id}
Test 4: List Payments

cmd
curl http://localhost:8080/api/v1/payments?page=1&limit=10
Test 5: Get Statistics

cmd
curl http://localhost:8080/api/v1/statistics
8.3 Sample Ethiopian Test Data
Create sample_data.json:

json
{
  "payments": [
    {
      "amount": 15000.75,
      "currency": "ETB",
      "reference": "CBE-FARMER-001",
      "description": "የአትክልት ምርት ክፍያ (Vegetable produce payment)",
      "customer_name": "አባ ለማ",
      "bank_code": "CBE"
    },
    {
      "amount": 500.00,
      "currency": "USD",
      "reference": "AWASH-EXPORT-001",
      "description": "Coffee export payment to Europe",
      "customer_name": "Addis Coffee Exporters",
      "bank_code": "AWASH"
    },
    {
      "amount": 2500.00,
      "currency": "ETB",
      "reference": "DASHEN-RENT-001",
      "description": "የንግድ ቤት ኪራይ (Shop rent payment)",
      "customer_name": "ፍቅር ንግድ",
      "bank_code": "DASHEN"
    },
    {
      "amount": 1200.50,
      "currency": "ETB",
      "reference": "ABYSSINIA-EDU-001",
      "description": "የትምህርት ክፍያ (University fee)",
      "customer_name": "ሰላም ታደሰ",
      "bank_code": "ABYSSINIA"
    },
    {
      "amount": 750.25,
      "currency": "USD",
      "reference": "NIB-IMPORT-001",
      "description": "Medical equipment import",
      "customer_name": "Addis Medical Supplies",
      "bank_code": "NIB"
    }
  ]
}
Test Concurrency and Idempotency
Test Idempotency:

cmd
# Send same payment twice
curl -X POST http://localhost:8080/api/v1/payments ^
  -H "Content-Type: application/json" ^
  -d "{\"amount\": 1000, \"currency\": \"ETB\", \"reference\": \"TEST-DUP-001\", \"description\": \"Test duplicate\"}"

# Same request again - should fail
curl -X POST http://localhost:8080/api/v1/payments ^
  -H "Content-Type: application/json" ^
  -d "{\"amount\": 1000, \"currency\": \"ETB\", \"reference\": \"TEST-DUP-001\", \"description\": \"Test duplicate\"}"
Test Concurrent Processing:

Create multiple payments quickly

Check RabbitMQ queue length

Watch worker logs processing concurrently
