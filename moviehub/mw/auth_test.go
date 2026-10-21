package mw

import(
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
 "moviehub/auth"
)
func TestProtectedRoutes(t *testing.T){
 iss:=&auth.Issuer{Secret:[]byte("0123456789abcdef0123456789abcdef"),TTL:time.Hour}
 next:=RequireLogin(iss,RequireRole("admin",func(w http.ResponseWriter,r *http.Request){w.WriteHeader(http.StatusNoContent)}))
 for _,tc:=range []struct{role string;want int}{{"member",403},{"admin",204}}{
  tok,err:=iss.Issue(99,tc.role);if err!=nil{t.Fatal(err)}
  req:=httptest.NewRequest("GET","/api/admin/stats",nil);req.Header.Set("Authorization","Bearer "+tok)
  w:=httptest.NewRecorder();next(w,req);if w.Code!=tc.want{t.Fatalf("role %s: got %d want %d",tc.role,w.Code,tc.want)}
 }
 w:=httptest.NewRecorder();next(w,httptest.NewRequest("GET","/api/admin/stats",nil))
 if w.Code!=401{t.Fatalf("missing token got %d",w.Code)}
}
