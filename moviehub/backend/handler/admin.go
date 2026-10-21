package handler

import (
 "encoding/json"
 "net/http"
 "moviehub/store"
)
type Admin struct{Store store.AdminStore}
func(h Admin)Stats(w http.ResponseWriter,r *http.Request){
 stats,e:=h.Store.Stats(r.Context())
 w.Header().Set("Content-Type","application/json")
 if e!=nil{w.WriteHeader(500);_ = json.NewEncoder(w).Encode(map[string]any{"error":map[string]string{"code":"DATABASE_ERROR","message":"อ่านสถิติไม่สำเร็จ"}});return}
 _=json.NewEncoder(w).Encode(stats)
}
