package mw

import (
 "net/http"
 "net/http/httptest"
 "testing"
)
func TestCORSPreflightMatchesWeek11(t *testing.T) {
 handler:=CORS([]string{"http://localhost:3000"},http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){t.Fatal("OPTIONS must not reach the handler")}))
 for _,tt:=range []struct{origin string;allowed bool}{
  {"http://localhost:3000",true},
  {"http://localhost:3001",false},
 } {
  request:=httptest.NewRequest("OPTIONS","/api/movies",nil)
  request.Header.Set("Origin",tt.origin)
  request.Header.Set("Access-Control-Request-Method","GET")
  w:=httptest.NewRecorder()
  handler.ServeHTTP(w,request)
  if w.Code!=http.StatusNoContent{t.Errorf("%s: got %d, want 204",tt.origin,w.Code)}
  if got:=w.Header().Get("Access-Control-Allow-Origin");tt.allowed&&got!=tt.origin || !tt.allowed&&got!=""{
   t.Errorf("%s: Allow-Origin=%q",tt.origin,got)
  }
  if !tt.allowed&&w.Header().Get("Access-Control-Allow-Methods")!=""{t.Error("disallowed origin received CORS methods")}
  if w.Header().Get("Vary")==""{t.Error("Vary Origin is required")}
 }
}
