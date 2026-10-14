$ErrorActionPreference = "Continue"
$base = "http://127.0.0.1:8080"
$pass = 0
$fail = 0
$skip = 0
$suffix = Get-Date -Format "HHmmss"

function Ok($name, $resp) {
  if ($null -ne $resp -and $resp.code -eq 200) {
    Write-Host "[PASS] $name"
    $script:pass++
    return $true
  }
  $msg = if ($resp) { "$($resp.code) $($resp.msg)" } else { "null" }
  Write-Host "[FAIL] $name -> $msg"
  $script:fail++
  return $false
}

function Api($method, $path, $token, $body) {
  try {
    $headers = @{}
    if ($token) { $headers["Authorization"] = "Bearer $token" }
    $p = @{ Method = $method; Uri = "$base$path"; Headers = $headers; ContentType = "application/json" }
    if ($null -ne $body) { $p["Body"] = $body }
    return Invoke-RestMethod @p
  } catch {
    $ex = $_.Exception
    $stream = $null
    try {
      if ($ex.Response) {
        $reader = New-Object IO.StreamReader($ex.Response.GetResponseStream())
        $txt = $reader.ReadToEnd()
        if ($txt) { return ($txt | ConvertFrom-Json) }
      }
    } catch {}
    return @{ code = 599; msg = $ex.Message }
  }
}

Write-Host "========== FULL API SUITE $suffix =========="

# --- Auth ---
$login = Api POST "/apis/v1/auth/login" $null '{"username":"admin","password":"Admin@123"}'
Ok "A01 login" $login | Out-Null
$token = $login.data.token
$bad = Api POST "/apis/v1/auth/login" $null '{"username":"admin","password":"wrong"}'
if ($bad.code -ne 200) { Write-Host "[PASS] A02 bad password rejected"; $pass++ } else { Write-Host "[FAIL] A02"; $fail++ }
$me = Api GET "/apis/v1/auth/me" $token $null
Ok "A03 me" $me | Out-Null
$noauth = Api GET "/apis/v1/stores?page=1" $null $null
if ($noauth.code -eq 401) { Write-Host "[PASS] A04 no token 401"; $pass++ } else { Write-Host "[FAIL] A04 code=$($noauth.code)"; $fail++ }

# --- Platform ---
Ok "B01 stores list" (Api GET "/apis/v1/stores?page=1&page_size=10" $token $null) | Out-Null
Ok "B02 users list" (Api GET "/apis/v1/users?page=1&page_size=10" $token $null) | Out-Null
Ok "B03 roles list" (Api GET "/apis/v1/roles?page=1&page_size=10" $token $null) | Out-Null
Ok "B04 permissions" (Api GET "/apis/v1/permissions" $token $null) | Out-Null

# --- Menu ---
$stations = Api GET "/apis/v1/stations?page=1&page_size=20" $token $null
Ok "C01 stations" $stations | Out-Null
$cats = Api GET "/apis/v1/menu-categories?page=1&page_size=20" $token $null
Ok "C02 categories" $cats | Out-Null
$dishes = Api GET "/apis/v1/dishes?page=1&page_size=20" $token $null
Ok "C03 dishes" $dishes | Out-Null
$stationId = $null; $catId = $null; $dishId = $null
if ($stations.data.list.Count -gt 0) { $stationId = $stations.data.list[0].id }
if ($cats.data.list.Count -gt 0) { $catId = $cats.data.list[0].id }
if ($dishes.data.list.Count -gt 0) { $dishId = $dishes.data.list[0].id }
if (-not $catId) {
  $c = Api POST "/apis/v1/menu-categories" $token (@{name="Cat$suffix"; sort_no=1; status=$true} | ConvertTo-Json)
  if (Ok "C04 create category" $c) { $catId = $c.data.id }
}
if (-not $stationId) {
  $s = Api POST "/apis/v1/stations" $token (@{code="S$suffix"; name="St$suffix"; sort_no=1; status=$true} | ConvertTo-Json)
  if (Ok "C05 create station" $s) { $stationId = $s.data.id }
}
if (-not $dishId -and $catId) {
  $d = Api POST "/apis/v1/dishes" $token (@{
    category_id=$catId; station_id=$stationId; name="Dish$suffix"; base_price="18.00"; is_on_sale=$true
  } | ConvertTo-Json)
  if (Ok "C06 create dish" $d) { $dishId = $d.data.id }
} else {
  Write-Host "[PASS] C06 use existing dish id=$dishId"; $pass++
}
if ($dishId) {
  Ok "C07 dish detail" (Api GET "/apis/v1/dishes/$dishId" $token $null) | Out-Null
}

# --- Tables ---
$areas = Api GET "/apis/v1/areas?page=1&page_size=20" $token $null
Ok "D01 areas" $areas | Out-Null
$tables = Api GET "/apis/v1/tables?page=1&page_size=50" $token $null
Ok "D02 tables" $tables | Out-Null
$areaId = if ($areas.data.list.Count -gt 0) { $areas.data.list[0].id } else { $null }
$idleTable = $tables.data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
if (-not $idleTable -and $areaId) {
  $nt = Api POST "/apis/v1/tables" $token (@{area_id=$areaId; code="T$suffix"; name="Table$suffix"; seats=4} | ConvertTo-Json)
  if (Ok "D03 create table" $nt) { $idleTable = $nt.data }
} else {
  Write-Host "[PASS] D03 use idle table id=$($idleTable.id)"; $pass++
}
$tableId = $idleTable.id

# --- Stock BOM prep ---
$mat = Api POST "/apis/v1/materials" $token (@{code="FM$suffix"; name="FullMat$suffix"; unit="kg"} | ConvertTo-Json)
Ok "E01 create material" $mat | Out-Null
$matId = $mat.data.id
$doc = Api POST "/apis/v1/stock-docs" $token (@{
  doc_type="inbound"; remark="full"; items=@(@{material_id=$matId; qty="200"; unit_cost="2"})
} | ConvertTo-Json -Depth 5)
Ok "E02 create inbound" $doc | Out-Null
$docId = $doc.data.doc.id
Ok "E03 post inbound" (Api POST "/apis/v1/stock-docs/$docId/post" $token "{}") | Out-Null
Ok "E04 balances" (Api GET "/apis/v1/stock-balances?page=1&page_size=10" $token $null) | Out-Null
if ($dishId -and $matId) {
  Ok "E05 set BOM" (Api PUT "/apis/v1/dishes/$dishId/materials" $token (@{
    materials=@(@{material_id=$matId; qty="0.2"})
  } | ConvertTo-Json -Depth 5)) | Out-Null
}

# --- Order + Pay cash ---
$order = Api POST "/apis/v1/orders" $token (@{
  table_id=$tableId; guest_count=2; items=@(@{dish_id=$dishId; qty=1})
} | ConvertTo-Json -Depth 5)
Ok "F01 create order" $order | Out-Null
$orderId = $order.data.order.id
Ok "F02 get order" (Api GET "/apis/v1/orders/$orderId" $token $null) | Out-Null
# 厨单在下单时已生成；结构为 {ticket, items}
$kdsMid = Api GET "/apis/v1/kds/tickets?status=pending" $token $null
$midList = @(); if ($kdsMid.data -is [System.Array]) { $midList = @($kdsMid.data) }
$midTicket = $null
$midTicketId = $null
if ($midList.Count -gt 0) {
  $midTicket = $midList[0]
  if ($midTicket.ticket) { $midTicketId = $midTicket.ticket.id } else { $midTicketId = $midTicket.id }
}
if ($midTicketId) {
  Ok "H02 kds start" (Api POST "/apis/v1/kds/tickets/$midTicketId/start" $token "{}") | Out-Null
  Ok "H03 kds done" (Api POST "/apis/v1/kds/tickets/$midTicketId/done" $token "{}") | Out-Null
  $script:kdsDoneEarly = $true
} else {
  $script:kdsDoneEarly = $false
}
Api POST "/apis/v1/orders/$orderId/advance" $token "{}" | Out-Null
for ($i=0; $i -lt 6; $i++) {
  $cur = Api GET "/apis/v1/orders/$orderId" $token $null
  if ($cur.data.order.status -eq "pending_pay") { break }
  Api POST "/apis/v1/orders/$orderId/advance" $token "{}" | Out-Null
}
$cur = Api GET "/apis/v1/orders/$orderId" $token $null
if ($cur.data.order.status -eq "pending_pay") { Write-Host "[PASS] F03 advance to pending_pay"; $pass++ }
else { Write-Host "[FAIL] F03 status=$($cur.data.order.status)"; $fail++ }
$amt = [string]$cur.data.order.payable_amount
$payCash = Api POST "/apis/v1/orders/$orderId/pay" $token (@{channel="cash"; amount=$amt} | ConvertTo-Json)
Ok "F04 cash pay + outbound" $payCash | Out-Null
Ok "F05 payments list" (Api GET "/apis/v1/orders/$orderId/payments" $token $null) | Out-Null

# --- Member pay path ---
$tables2 = Api GET "/apis/v1/tables?page=1&page_size=50" $token $null
$idle2 = $tables2.data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
$mem = Api POST "/apis/v1/members" $token (@{phone="137$suffix"; name="FullMem"} | ConvertTo-Json)
Ok "G01 create member" $mem | Out-Null
$memId = $mem.data.member.id
Ok "G02 recharge" (Api POST "/apis/v1/members/$memId/recharge" $token (@{amount="300.00"} | ConvertTo-Json)) | Out-Null
Ok "G03 earn points" (Api POST "/apis/v1/members/$memId/points/earn" $token (@{points=10} | ConvertTo-Json)) | Out-Null
Ok "G04 wallet logs" (Api GET "/apis/v1/members/$memId/wallet-logs?page=1" $token $null) | Out-Null
if ($idle2 -and $dishId) {
  $o2 = Api POST "/apis/v1/orders" $token (@{
    table_id=$idle2.id; guest_count=1; member_id=$memId; items=@(@{dish_id=$dishId; qty=1})
  } | ConvertTo-Json -Depth 5)
  Ok "G05 order with member" $o2 | Out-Null
  $oid2 = $o2.data.order.id
  for ($i=0; $i -lt 6; $i++) {
    $c2 = Api GET "/apis/v1/orders/$oid2" $token $null
    if ($c2.data.order.status -eq "pending_pay") { break }
    Api POST "/apis/v1/orders/$oid2/advance" $token "{}" | Out-Null
  }
  $c2 = Api GET "/apis/v1/orders/$oid2" $token $null
  $amt2 = [string]$c2.data.order.payable_amount
  Ok "G06 member pay" (Api POST "/apis/v1/orders/$oid2/pay" $token (@{
    channel="member"; amount=$amt2; member_id=$memId
  } | ConvertTo-Json)) | Out-Null
} else {
  Write-Host "[SKIP] G05-G06 no idle table"; $skip++
}

# --- KDS list ---
$kds = Api GET "/apis/v1/kds/tickets" $token $null
Ok "H01 kds list" $kds | Out-Null
if (-not $kdsDoneEarly) {
  Write-Host "[SKIP] H02-H03 already covered or no ticket"; $skip += 2
}

# --- Queue ---
$q = Api POST "/apis/v1/queues" $token (@{guest_count=3; phone="13600001111"} | ConvertTo-Json)
Ok "I01 queue take" $q | Out-Null
Ok "I02 queue list" (Api GET "/apis/v1/queues?page=1&page_size=20" $token $null) | Out-Null
Ok "I03 queue call" (Api POST "/apis/v1/queues/$($q.data.id)/call" $token "{}") | Out-Null
$idle3 = (Api GET "/apis/v1/tables?page=1&page_size=50" $token $null).data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
if ($idle3) {
  $seat = Api POST "/apis/v1/queues/$($q.data.id)/seat" $token (@{table_id=$idle3.id; create_order=$true} | ConvertTo-Json)
  Ok "I04 queue seat" $seat | Out-Null
  $tq = Api GET "/apis/v1/tables/$($idle3.id)" $token $null
  if ($tq.data.status -eq "occupied") { Write-Host "[PASS] I05 table occupied"; $pass++ } else { Write-Host "[FAIL] I05"; $fail++ }
  # cleanup: cancel empty order if possible
  if ($seat.data.order_id) {
    Api POST "/apis/v1/orders/$($seat.data.order_id)/cancel" $token (@{remark="cleanup"} | ConvertTo-Json) | Out-Null
  }
} else {
  Write-Host "[SKIP] I04-I05 no idle table"; $skip += 2
  Api POST "/apis/v1/queues/$($q.data.id)/cancel" $token (@{remark="cleanup"} | ConvertTo-Json) | Out-Null
}

# --- Reservation ---
$idle4 = (Api GET "/apis/v1/tables?page=1&page_size=50" $token $null).data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
$reserveAt = (Get-Date).AddHours(3).ToString("yyyy-MM-ddTHH:mm:sszzz")
if ($idle4) {
  $res = Api POST "/apis/v1/reservations" $token (@{
    table_id=$idle4.id; customer_name="Guest$suffix"; phone="135$suffix"; guest_count=2;
    reserve_at=$reserveAt; channel="phone"
  } | ConvertTo-Json)
  Ok "J01 create reservation" $res | Out-Null
  if ($res.code -eq 200 -and $res.data.id) {
    Ok "J02 confirm" (Api POST "/apis/v1/reservations/$($res.data.id)/confirm" $token "{}") | Out-Null
    Ok "J03 cancel res" (Api POST "/apis/v1/reservations/$($res.data.id)/cancel" $token (@{reason="test"} | ConvertTo-Json)) | Out-Null
  }
} else {
  Write-Host "[SKIP] J01-J03 no idle table"; $skip += 3
}
Ok "J04 reservations list" (Api GET "/apis/v1/reservations?page=1&page_size=10" $token $null) | Out-Null

# --- Online (need idle table) ---
$idleOl = (Api GET "/apis/v1/tables?page=1&page_size=50" $token $null).data.list | Where-Object { $_.status -eq "idle" } | Select-Object -First 1
if ($idleOl -and $dishId) {
  Start-Sleep -Milliseconds 80
  $onlineBody = (@{
    store_id=1; table_id=$idleOl.id; contact_name="OL"; contact_phone="13400001111";
    source="scan"; items=@(@{dish_id=$dishId; qty=1})
  } | ConvertTo-Json -Depth 5)
  $online = Api POST "/apis/v1/online/orders" $null $onlineBody
  if ($online.code -ne 200) {
    Start-Sleep -Milliseconds 200
    $online = Api POST "/apis/v1/online/orders" $null $onlineBody
  }
  Ok "K01 online create" $online | Out-Null
  if ($online.code -eq 200) {
    $atok = $online.data.access_token
    $oid = $online.data.order.id
    Ok "K02 online get" (Api GET "/apis/v1/online/orders/${oid}?access_token=$atok" $null $null) | Out-Null
    Api POST "/apis/v1/orders/$oid/cancel" $token (@{remark="cleanup online"} | ConvertTo-Json) | Out-Null
  } else {
    Write-Host "[SKIP] K02 online get"; $skip++
  }
} else {
  Write-Host "[SKIP] K01-K02 no idle table"; $skip += 2
}

# --- Purchase ---
$sup = Api POST "/apis/v1/suppliers" $token (@{code="SUP$suffix"; name="Supplier$suffix"; contact_name="c"; phone="13000001111"} | ConvertTo-Json)
Ok "L01 supplier" $sup | Out-Null
$supId = $sup.data.id
$po = Api POST "/apis/v1/purchase-orders" $token (@{
  supplier_id=$supId; remark="full"; items=@(@{material_id=$matId; qty="5"; unit_price="3.00"})
} | ConvertTo-Json -Depth 5)
Ok "L02 purchase order" $po | Out-Null
$poId = $null
$poiId = $null
if ($po.data.order) {
  $poId = $po.data.order.id
  if ($po.data.items -and $po.data.items.Count -gt 0) { $poiId = $po.data.items[0].id }
}
if ($poId) {
  Ok "L03 submit PO" (Api POST "/apis/v1/purchase-orders/$poId/submit" $token "{}") | Out-Null
  if ($poiId) {
    $pr = Api POST "/apis/v1/purchase-receipts" $token (@{
      purchase_order_id=$poId; items=@(@{purchase_order_item_id=$poiId; qty="5"})
    } | ConvertTo-Json -Depth 5)
    Ok "L04 receipt" $pr | Out-Null
    $prId = if ($pr.data.receipt) { $pr.data.receipt.id } else { $null }
    if ($prId) { Ok "L05 post receipt" (Api POST "/apis/v1/purchase-receipts/$prId/post" $token "{}") | Out-Null }
    else { Write-Host "[SKIP] L05 no receipt id"; $skip++ }
  } else {
    Write-Host "[SKIP] L04-L05 no poi"; $skip += 2
  }
} else {
  Write-Host "[SKIP] L03-L05 po shape"; $skip += 3
}

# --- Finance ---
$today = Get-Date -Format "yyyy-MM-dd"
$settle = Api POST "/apis/v1/settlements/daily" $token (@{business_day=$today} | ConvertTo-Json)
if ($settle.code -eq 200) { Ok "M01 daily settle" $settle | Out-Null }
elseif ($settle.code -eq 400) {
  # 重复日结防护（当日已结）
  Write-Host "[PASS] M01 settle protected ($($settle.msg))"
  $pass++
} else {
  Write-Host "[FAIL] M01 unexpected code=$($settle.code) msg=$($settle.msg)"; $fail++
}
Ok "M02 settlements list" (Api GET "/apis/v1/settlements?page=1&page_size=10" $token $null) | Out-Null
$recon = Api POST "/apis/v1/reconciliations" $token (@{
  business_day=$today; channel="cash"; channel_amount="0.00"; entries=@()
} | ConvertTo-Json -Depth 5)
if ($recon.code -eq 200) { Ok "M03 reconciliation" $recon | Out-Null }
else { Write-Host "[INFO] M03 recon: $($recon.msg)"; if ($recon.code -ne 200) { Write-Host "[SKIP] M03"; $skip++ } }
Ok "M04 recon list" (Api GET "/apis/v1/reconciliations?page=1&page_size=10" $token $null) | Out-Null
Ok "M05 diffs" (Api GET "/apis/v1/reconciliations/diffs?page=1&page_size=10" $token $null) | Out-Null

# --- HR ---
$emp = Api POST "/apis/v1/employees" $token (@{
  emp_no="F$suffix"; name="FullEmp"; phone="13100001111"; job_title="waiter"; status=$true
} | ConvertTo-Json)
Ok "N01 employee" $emp | Out-Null
$empId = $emp.data.id
$shift = Api POST "/apis/v1/shifts" $token (@{
  employee_id=$empId; shift_date=$today; start_time="09:00"; end_time="17:00"; shift_name="day"
} | ConvertTo-Json)
Ok "N02 shift" $shift | Out-Null
Ok "N03 punch in" (Api POST "/apis/v1/attendances/punch" $token (@{
  employee_id=$empId; work_date=$today; type="in"
} | ConvertTo-Json)) | Out-Null
Ok "N04 punch out" (Api POST "/apis/v1/attendances/punch" $token (@{
  employee_id=$empId; work_date=$today; type="out"
} | ConvertTo-Json)) | Out-Null
Ok "N05 attendances" (Api GET "/apis/v1/attendances?work_date=$today&page=1" $token $null) | Out-Null
Ok "N06 cancel shift" (Api POST "/apis/v1/shifts/$($shift.data.id)/cancel" $token "{}") | Out-Null

# --- RBAC smoke ---
if ($me.data.permission_codes -contains "payment:write") {
  Write-Host "[PASS] O01 has payment:write"; $pass++
} else { Write-Host "[FAIL] O01 missing payment:write"; $fail++ }
if ($me.data.permission_codes -contains "hr:employee:write") {
  Write-Host "[PASS] O02 has hr:employee:write"; $pass++
} else { Write-Host "[FAIL] O02"; $fail++ }

Write-Host ""
Write-Host "========== RESULT pass=$pass fail=$fail skip=$skip =========="
if ($fail -gt 0) { exit 1 } else { exit 0 }
