package main
import("testing";"net/http/httptest")
func TestPrivateMobileGate(t *testing.T){
 t.Setenv("INKVEIL_PRIVATE_MOBILE","yes");t.Setenv("INKVEIL_MOBILE_PASSWORD","test-password-long-enough")
 h:=configuredHandler("");w:=httptest.NewRecorder();h.ServeHTTP(w,httptest.NewRequest("GET","/health",nil));if w.Code!=401{t.Fatal(w.Code)}
 r:=httptest.NewRequest("GET","/health",nil);r.SetBasicAuth("inkveil","test-password-long-enough");w=httptest.NewRecorder();h.ServeHTTP(w,r);if w.Code!=200{t.Fatal(w.Code)}
}
