package mw

import (
 "context"
 "encoding/json"
 "net/http"
 "strings"
 "moviehub/auth"
)
type ctxKey int
const(memberIDKey ctxKey=iota;roleKey)
func MemberID(ctx context.Context)(int64,bool){id,ok:=ctx.Value(memberIDKey).(int64);return id,ok}
func Role(ctx context.Context)string{v,_:=ctx.Value(roleKey).(string);return v}
func deny(w http.ResponseWriter,status int,code,msg string){
 w.Header().Set("Content-Type","application/json")
 w.WriteHeader(status)
 _=json.NewEncoder(w).Encode(map[string]any{"error":map[string]string{"code":code,"message":msg}})
}
func RequireLogin(iss *auth.Issuer,next http.HandlerFunc)http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  parts:=strings.Fields(r.Header.Get("Authorization"))
  if len(parts)!=2||!strings.EqualFold(parts[0],"Bearer"){deny(w,401,"NO_TOKEN","กรุณาเข้าสู่ระบบก่อน");return}
  id,role,err:=iss.Parse(parts[1])
  if err!=nil||id<=0{deny(w,401,"BAD_TOKEN","ตั๋วไม่ถูกต้องหรือหมดอายุ");return}
  ctx:=context.WithValue(r.Context(),memberIDKey,id)
  ctx=context.WithValue(ctx,roleKey,role)
  next(w,r.WithContext(ctx))
 }
}
func RequireRole(role string,next http.HandlerFunc)http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  if Role(r.Context())!=role{deny(w,403,"FORBIDDEN","ต้องเป็น "+role+" เท่านั้น");return}
  next(w,r)
 }
}
