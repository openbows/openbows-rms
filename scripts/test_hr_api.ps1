$ErrorActionPreference = "Stop"
$base = "http://127.0.0.1:8080"
$pass = 0
$fail = 0

function Assert-Ok($name, $resp) {
  if ($resp.code -eq 200) {
    Write-Host "[PASS] $name"
    $script:pass++
  } else {
    Write-Host "[FAIL] $name code=$($resp.code) msg=$($resp.msg)"
    $script:fail++
  }
}

function Api($method, $path, $token, $body) {
  $headers = @{}
  if ($token) { $headers["Authorization"] = "Bearer $token" }
  $params = @{
    Method = $method
    Uri = "$base$path"
    Headers = $headers
    ContentType = "application/json"
  }
  if ($null -ne $body) { $params["Body"] = $body }
  return Invoke-RestMethod @params
}

Write-Host "=== HR API Test ==="

$login = Api POST "/apis/v1/auth/login" $null '{"username":"admin","password":"Admin@123"}'
Assert-Ok "login" $login
$token = $login.data.token
if (-not $token) {
  Write-Host "FATAL: no token"
  exit 1
}

# probe route
try {
  $probe = Api GET "/apis/v1/employees?page=1&page_size=5" $token $null
  Assert-Ok "list employees (route)" $probe
} catch {
  Write-Host "[FAIL] employees route missing: $($_.Exception.Message)"
  Write-Host "HINT: restart server with new binary"
  exit 2
}

$suffix = Get-Date -Format "HHmmss"
$empBody = @{
  emp_no = "E$suffix"
  name = "Test Emp $suffix"
  phone = "13800138000"
  job_title = "cashier"
  status = $true
  hired_at = (Get-Date -Format "yyyy-MM-dd")
  remark = "hr api test"
} | ConvertTo-Json -Compress

$emp = Api POST "/apis/v1/employees" $token $empBody
Assert-Ok "create employee" $emp
$empId = $emp.data.id
Write-Host "  emp_id=$empId emp_no=$($emp.data.emp_no)"

$got = Api GET "/apis/v1/employees/$empId" $token $null
Assert-Ok "get employee" $got

$updBody = @{
  emp_no = "E$suffix"
  name = "Test Emp Updated"
  phone = "13900139000"
  job_title = "waiter"
  status = $true
  hired_at = (Get-Date -Format "yyyy-MM-dd")
  remark = "updated"
} | ConvertTo-Json -Compress
$upd = Api PUT "/apis/v1/employees/$empId" $token $updBody
Assert-Ok "update employee" $upd

$list = Api GET "/apis/v1/employees?page=1&page_size=10&keyword=E$suffix" $token $null
Assert-Ok "list employees keyword" $list

$today = Get-Date -Format "yyyy-MM-dd"
$shiftBody = @{
  employee_id = $empId
  shift_date = $today
  start_time = "09:00"
  end_time = "17:00"
  shift_name = "day"
  remark = "shift test"
} | ConvertTo-Json -Compress
$shift = Api POST "/apis/v1/shifts" $token $shiftBody
Assert-Ok "create shift" $shift
$shiftId = $shift.data.id
Write-Host "  shift_id=$shiftId"

$shiftGet = Api GET "/apis/v1/shifts/$shiftId" $token $null
Assert-Ok "get shift" $shiftGet

$shiftList = Api GET "/apis/v1/shifts?page=1&page_size=10&shift_date=$today&employee_id=$empId" $token $null
Assert-Ok "list shifts by day" $shiftList

$shiftUpd = @{
  employee_id = $empId
  shift_date = $today
  start_time = "10:00:00"
  end_time = "18:00:00"
  shift_name = "day2"
  remark = "shift updated"
} | ConvertTo-Json -Compress
$su = Api PUT "/apis/v1/shifts/$shiftId" $token $shiftUpd
Assert-Ok "update shift" $su

$punchIn = @{
  employee_id = $empId
  work_date = $today
  type = "in"
  remark = "punch in"
} | ConvertTo-Json -Compress
$pin = Api POST "/apis/v1/attendances/punch" $token $punchIn
Assert-Ok "punch in" $pin
$attId = $pin.data.id
Write-Host "  attendance_id=$attId clock_in=$($pin.data.clock_in_at)"

$punchOut = @{
  employee_id = $empId
  work_date = $today
  type = "out"
} | ConvertTo-Json -Compress
$pout = Api POST "/apis/v1/attendances/punch" $token $punchOut
Assert-Ok "punch out" $pout

$attList = Api GET "/apis/v1/attendances?page=1&page_size=10&work_date=$today&employee_id=$empId" $token $null
Assert-Ok "list attendances by day" $attList

$yesterday = (Get-Date).AddDays(-1).ToString("yyyy-MM-dd")
$makeupBody = @{
  employee_id = $empId
  work_date = $yesterday
  clock_in_at = "$yesterday 09:05:00"
  clock_out_at = "$yesterday 18:10:00"
  remark = "makeup"
} | ConvertTo-Json -Compress
$mk = Api POST "/apis/v1/attendances/makeup" $token $makeupBody
Assert-Ok "makeup" $mk

$attGet = Api GET "/apis/v1/attendances/$attId" $token $null
Assert-Ok "get attendance" $attGet

$cancel = Api POST "/apis/v1/shifts/$shiftId/cancel" $token "{}"
Assert-Ok "cancel shift" $cancel

# negative: duplicate punch in
try {
  $dup = Api POST "/apis/v1/attendances/punch" $token $punchIn
  if ($dup.code -ne 200) {
    Write-Host "[PASS] duplicate punch rejected code=$($dup.code)"
    $pass++
  } else {
    Write-Host "[FAIL] duplicate punch should fail"
    $fail++
  }
} catch {
  Write-Host "[PASS] duplicate punch rejected (http err)"
  $pass++
}

Write-Host ""
Write-Host "=== RESULT pass=$pass fail=$fail ==="
if ($fail -gt 0) { exit 1 } else { exit 0 }
