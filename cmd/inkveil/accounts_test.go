package main
import("testing";"net/http/httptest";"golang.org/x/crypto/bcrypt")
func TestAccountDisabledWithoutDB(t *testing.T){if catalogDB!=nil{t.Skip("database active")};w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/account",nil));if w.Code!=503{t.Fatal(w.Code)}}
func TestAccountTokenAndPassword(t *testing.T){a,e:=randomToken();if e!=nil||len(a)!=64||len(tokenHash(a))!=32{t.Fatal("token")};b,e:=randomToken();if e!=nil||a==b{t.Fatal("random token")};h,e:=bcrypt.GenerateFromPassword([]byte("example-password"),bcrypt.MinCost);if e!=nil||bcrypt.CompareHashAndPassword(h,[]byte("example-password"))!=nil{t.Fatal("hash")};if accountName.MatchString("bad/name"){t.Fatal("invalid username")}}
