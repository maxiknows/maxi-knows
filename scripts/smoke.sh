#!/usr/bin/env bash
# Usage: BASE=http://localhost:8080 [FULL=1] bash scripts/smoke.sh
#   FULL=1  also tests register/login/logout (creates users; never use on prod)
set -u
BASE="${BASE:-http://localhost:8080}"
FULL="${FULL:-0}"
STATIC_PATH="${STATIC_PATH:-/static/style.css}"   # adjust to a real file
FAILS=0

code() { curl -s -o /dev/null -m 15 -w '%{http_code}' "$@"; }

check() {
  if [ "$2" = "$3" ]; then echo "PASS $1"
  else echo "FAIL $1 (expected $2, got $3)"; FAILS=$((FAILS+1)); fi
}

# Upstream APIs (Open-Meteo, Nominatim) can be flaky; warn instead of failing.
soft() {
  if [ "$2" = "$3" ]; then echo "PASS $1"
  else echo "WARN $1 (expected $2, got $3)"; fi
}

# --- pages ---
check "GET /"         200 "$(code "$BASE/")"
check "GET /register" 200 "$(code "$BASE/register")"
check "GET /login"    200 "$(code "$BASE/login")"
check "GET /about"    200 "$(code "$BASE/about")"
check "static file"   200 "$(code "$BASE$STATIC_PATH")"

# --- search ---
check "api/search ok"        200 "$(code "$BASE/api/search?q=test")"
check "api/search missing q" 422 "$(code "$BASE/api/search")"
check "search page w/ query" 200 "$(code "$BASE/?q=test")"

# --- weather / location ---
check "weather bad lat"   400 "$(code "$BASE/api/weather?lat=abc&lon=12")"
check "weather range"     400 "$(code "$BASE/api/weather?lat=91&lon=12")"
check "location missing"  400 "$(code "$BASE/api/location")"
soft  "weather upstream"  200 "$(code "$BASE/api/weather?lat=55.68&lon=12.57")"
soft  "location upstream" 200 "$(code "$BASE/api/location?lat=55.68&lon=12.57")"

# --- auth flow (creates data) ---
if [ "$FULL" = "1" ]; then
  JAR=$(mktemp); OLD=$(mktemp)
  U="smoke$RANDOM"; PW="pw12345"

  check "register missing fields" 422 "$(code -X POST -d "username=x" "$BASE/api/register")"
  check "register bad email"      422 "$(code -X POST -d "username=$U&email=nope&password=$PW&password2=$PW" "$BASE/api/register")"
  check "register pw mismatch"    422 "$(code -X POST -d "username=$U&email=$U@x.dk&password=$PW&password2=other" "$BASE/api/register")"
  check "register ok"             303 "$(code -X POST -d "username=$U&email=$U@x.dk&password=$PW&password2=$PW" "$BASE/api/register")"
  check "register duplicate"      409 "$(code -X POST -d "username=$U&email=$U@x.dk&password=$PW&password2=$PW" "$BASE/api/register")"

  check "login unknown user" 401 "$(code -X POST -d "username=nobody$RANDOM&password=$PW" "$BASE/api/login")"
  check "login wrong pw"     401 "$(code -X POST -d "username=$U&password=wrong" "$BASE/api/login")"
  check "login ok"           303 "$(code -c "$JAR" -X POST -d "username=$U&password=$PW" "$BASE/api/login")"

  if curl -s -m 15 -b "$JAR" "$BASE/" | grep -q "$U"; then echo "PASS username shown when logged in"
  else echo "FAIL username not shown (cookie dropped?)"; FAILS=$((FAILS+1)); fi
  cp "$JAR" "$OLD"

  check "logout" 303 "$(code -b "$JAR" -c "$JAR" -X POST "$BASE/api/logout")"
  if curl -s -m 15 -b "$OLD" "$BASE/" | grep -q "$U"; then
    echo "FAIL old session still valid after logout"; FAILS=$((FAILS+1))
  else echo "PASS old session invalid after logout"; fi
fi

echo
if [ "$FAILS" -gt 0 ]; then echo "$FAILS check(s) failed"; exit 1; fi
echo "All checks passed"