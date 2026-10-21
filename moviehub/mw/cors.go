package mw

import (
 "net/http"
 "strings"
)
func CORS(origins []string,next http.Handler)http.Handler{
 allowed:=map[string]bool{}
 for _,o:=range origins{o=strings.TrimSpace(o);if o!=""{allowed[o]=true}}
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  origin:=r.Header.Get("Origin")
  w.Header().Add("Vary","Origin")
  if origin!=""&&allowed[origin]{
   w.Header().Set("Access-Control-Allow-Origin",origin)
   w.Header().Set("Access-Control-Allow-Methods","GET, POST, PUT, PATCH, DELETE, OPTIONS")
   w.Header().Set("Access-Control-Allow-Headers","Authorization, Content-Type")
  }
  if r.Method==http.MethodOptions {
   w.Header().Add("Vary","Access-Control-Request-Method")
   w.Header().Add("Vary","Access-Control-Request-Headers")
   w.WriteHeader(http.StatusNoContent);return
  }
  next.ServeHTTP(w,r)
 })
}
