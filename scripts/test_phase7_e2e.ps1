$ErrorActionPreference = "Stop"
$base = "http://127.0.0.1:8080"
$pass = 0
$fail = 0

function Assert-Ok($name, $resp) {
  if ($resp.code -eq 200) { Write-Host "[PASS] $name"; $script:pass++ }
  else { Write-Host "[FAIL] $name code=$($resp.code) msg=$($resp.msg)"; $script:fail++ }
}

function Api($method, $path, $token, $body) {
  $headers = @{}
  if ($token) { $headers["Authorization"] = "Bearer $token" }
  $params = @{ Method=$method; Uri="$base$path"; Headers=$headers; ContentType="application/json" }
  if ($null -ne $body) { $params["Body"] = $body }
  return Invoke-RestMethod @params
}

Write-Host "=== Phase7 E2E ==="
$login = Api POST "/apis/v1/auth/login" $null '{"username":"admin","password":"Admin@123"}'
Assert-Ok "login" $login
$token = $login.data.token
$suffix = Get-Date -Format "HHmmss"

# material + inbound stock
$mat = Api POST "/apis/v1/materials" $token (@{code="M$suffix"; name="Mat $suffix"; unit="kg"} | ConvertTo-Json)
Assert-Ok "create material" $mat
$matId = $mat.data.id

$doc = Api POST "/apis/v1/stock-docs" $token (@{
  doc_type="inbound"; remark="seed"; items=@(@{material_id=$matId; qty="100"; unit_cost="1"})
} | ConvertTo-Json -Depth 5)
Assert-Ok "create inbound" $doc
$posted = Api POST "/apis/v1/stock-docs/$($doc.data.doc.id)/post" $token "{}"
Assert-Ok "post inbound" $posted

# need a dish - list first
$dishes = Api GET "/apis/v1/dishes?page=1&page_size=1" $token $null
Assert-Ok "list dishes" $dishes
if (-not $dishes.data.list -or $dishes.data.list.Count -eq 0) {
  Write-Host "[FAIL] no dish for bom test"; $fail++; Write-Host "RESULT pass=$pass fail=$fail"; exit 1
}
$dishId = $dishes.data.list[0].id
$bom = Api PUT "/apis/v1/dishes/$dishId/materials" $token (@{
  materials=@(@{material_id=$matId; qty="0.5"})
} | ConvertTo-Json -Depth 5)
Assert-Ok "set dish bom" $bom

# tables
$tables = Api GET "/apis/v1/tables?page=1&page_size=20" $token $null
Assert-Ok "list tables" $tables
$idle = $tables.data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
if (-not $idle) { Write-Host "[FAIL] no idle table"; $fail++; exit 1 }
$tableId = $idle.id

# create order
$order = Api POST "/apis/v1/orders" $token (@{
  table_id=$tableId; guest_count=2; items=@(@{dish_id=$dishId; qty=2})
} | ConvertTo-Json -Depth 5)
Assert-Ok "create order" $order
$orderId = $order.data.order.id

# advance to pending_pay (may need multiple)
for ($i=0; $i -lt 5; $i++) {
  $cur = Api GET "/apis/v1/orders/$orderId" $token $null
  $st = $cur.data.order.status
  if ($st -eq "pending_pay") { break }
  if ($st -eq "completed" -or $st -eq "cancelled") { break }
  try { Api POST "/apis/v1/orders/$orderId/advance" $token "{}" | Out-Null } catch { break }
}
$cur = Api GET "/apis/v1/orders/$orderId" $token $null
Write-Host "  order status=$($cur.data.order.status) payable=$($cur.data.order.payable_amount)"
if ($cur.data.order.status -ne "pending_pay") {
  Write-Host "[FAIL] order not pending_pay"; $fail++
} else {
  Write-Host "[PASS] order pending_pay"; $pass++
  $amt = $cur.data.order.payable_amount
  # ensure string
  if ($amt -isnot [string]) { $amt = [string]$amt }

  # member for pay
  $mem = Api POST "/apis/v1/members" $token (@{phone="139$suffix"; name="E2E Mem"} | ConvertTo-Json)
  Assert-Ok "create member" $mem
  $memId = $mem.data.member.id
  $rech = Api POST "/apis/v1/members/$memId/recharge" $token (@{amount="500.00"; remark="e2e"} | ConvertTo-Json)
  Assert-Ok "recharge member" $rech

  $pay = Api POST "/apis/v1/orders/$orderId/pay" $token (@{
    channel="member"; amount=$amt; member_id=$memId
  } | ConvertTo-Json)
  Assert-Ok "member pay + stock out" $pay

  $bal = Api GET "/apis/v1/stock-balances?page=1&page_size=50" $token $null
  Assert-Ok "list balances" $bal
  $row = $bal.data.list | Where-Object { $_.balance.material_id -eq $matId } | Select-Object -First 1
  if ($row -and [decimal]$row.balance.qty -lt 100) {
    Write-Host "[PASS] stock deducted qty=$($row.balance.qty)"; $pass++
  } else {
    Write-Host "[FAIL] stock not deducted qty=$($row.balance.qty)"; $fail++
  }
}

# queue seat
$q = Api POST "/apis/v1/queues" $token (@{guest_count=2; phone="13800001111"} | ConvertTo-Json)
Assert-Ok "queue take" $q
$idle2 = (Api GET "/apis/v1/tables?page=1&page_size=50" $token $null).data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
if ($idle2) {
  $seat = Api POST "/apis/v1/queues/$($q.data.id)/seat" $token (@{table_id=$idle2.id; create_order=$true} | ConvertTo-Json)
  Assert-Ok "queue seat create order" $seat
  $t2 = Api GET "/apis/v1/tables/$($idle2.id)" $token $null
  if ($t2.data.status -eq "occupied") { Write-Host "[PASS] table occupied after seat"; $pass++ }
  else { Write-Host "[FAIL] table not occupied"; $fail++ }
}

# rbac: me has permissions
$me = Api GET "/apis/v1/auth/me" $token $null
Assert-Ok "auth me" $me
if ($me.data.permission_codes -contains "order:write") {
  Write-Host "[PASS] permission codes in me"; $pass++
} else {
  Write-Host "[FAIL] missing order:write in me"; $fail++
}

Write-Host ""
Write-Host "=== RESULT pass=$pass fail=$fail ==="
if ($fail -gt 0) { exit 1 } else { exit 0 }
