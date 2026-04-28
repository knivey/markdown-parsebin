#!/bin/bash
set -e

DB="/tmp/test-api-full.db"
rm -f "$DB"
BIN="./dave-web"

echo "=== Starting web server ==="
$BIN serve --db "$DB" --addr :18080 &
SRV_PID=$!
sleep 1

echo "=== Test 1: No API key (expect 401) ==="
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:18080/api/pastes \
  -H "Content-Type: application/json" -d '{"content":"test"}')
echo "Status: $HTTP"
[ "$HTTP" = "401" ] && echo "PASS" || echo "FAIL"

echo "=== Test 2: Bad API key (expect 401) ==="
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:18080/api/pastes \
  -H "Content-Type: application/json" -H "X-API-Key: badkey" -d '{"content":"test"}')
echo "Status: $HTTP"
[ "$HTTP" = "401" ] && echo "PASS" || echo "FAIL"

echo "=== Creating API key ==="
KEY=$($BIN --db "$DB" keys create --description "test" 2>/dev/null)
echo "Key: ${KEY:0:16}..."

echo "=== Test 3: Valid key (expect 201) ==="
RESP=$(curl -s -w "\n%{http_code}" -X POST http://localhost:18080/api/pastes \
  -H "Content-Type: application/json" -H "X-API-Key: $KEY" \
  -d '{"content":"# Hello\n\n**world**","title":"Test Paste"}')
HTTP=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -1)
echo "Status: $HTTP"
echo "Body: $BODY"
[ "$HTTP" = "201" ] && echo "PASS" || echo "FAIL"

echo "=== Test 4: View paste in browser ==="
PASTE_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
RESP=$(curl -s -w "\n%{http_code}" "http://localhost:18080/p/$PASTE_ID")
HTTP=$(echo "$RESP" | tail -1)
echo "Status: $HTTP"
[ "$HTTP" = "200" ] && echo "PASS" || echo "FAIL"

echo "=== Test 5: List keys ==="
$BIN --db "$DB" keys list 2>/dev/null

echo "=== Test 6: Revoke key then retry (expect 401) ==="
$BIN --db "$DB" keys revoke --key "$KEY" 2>/dev/null
HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:18080/api/pastes \
  -H "Content-Type: application/json" -H "X-API-Key: $KEY" -d '{"content":"test"}')
echo "Status: $HTTP"
[ "$HTTP" = "401" ] && echo "PASS" || echo "FAIL"

echo "=== Done ==="
kill $SRV_PID 2>/dev/null
