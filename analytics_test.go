package main
import("testing";"time";"net/http/httptest")
func TestAnalyticsDayWIB(t *testing.T){if dayKey(time.Date(2026,11,26,17,0,0,0,time.UTC))!=20261127{t.Fatal("WIB boundary")}}
func TestRankingPeriods(t *testing.T){for p,n:=range map[string]int{"today":1,"weekly":7,"monthly":30,"invalid":7}{_,got:=rankingPeriod(p);if got!=n{t.Fatal(p,got)}}}
func TestRankingWithoutDatabase(t *testing.T){if catalogDB!=nil{t.Skip("database active")};w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/ranking",nil));if w.Code!=200{t.Fatal(w.Code)}}
