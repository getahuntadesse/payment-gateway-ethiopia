@echo off
echo ==========================================
echo   Ethiopian Payment Gateway Test
echo ==========================================
echo.

echo Testing API endpoints...
echo.

REM Test 1: Health Check
echo 1. Testing Health Check...
curl -s http://localhost:8080/api/v1/health | python -m json.tool
if errorlevel 1 (
    echo ERROR: API is not responding
    echo Make sure API is running: go run cmd/api/main.go
    goto :end
)
echo.

REM Test 2: Create Payment
echo 2. Creating a test payment...
set timestamp=%date:~10,4%%date:~4,2%%date:~7,2%%time:~0,2%%time:~3,2%%time:~6,2%
set timestamp=%timestamp: =0%
curl -X POST http://localhost:8080/api/v1/payments ^
  -H "Content-Type: application/json" ^
  -d "{\"amount\": 5000.50, \"currency\": \"ETB\", \"reference\": \"TEST-%timestamp%\", \"description\": \"Test payment\", \"customer_name\": \"Test Customer\", \"bank_code\": \"CBE\"}" ^
  | python -m json.tool
echo.

REM Test 3: List Payments
echo 3. Listing payments...
curl -s http://localhost:8080/api/v1/payments?limit=5 | python -m json.tool
echo.

REM Test 4: Get Statistics
echo 4. Getting statistics...
curl -s http://localhost:8080/api/v1/statistics | python -m json.tool
echo.

:end
echo ==========================================
echo   Test Complete
echo ==========================================
echo.
echo Next steps:
echo 1. Check worker terminal for payment processing
echo 2. Open RabbitMQ: http://localhost:15672
echo 3. Run test again in 30 seconds
pause