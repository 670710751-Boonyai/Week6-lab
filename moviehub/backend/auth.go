package main

import (
 "context"
 "crypto/hmac"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "errors"
 "net/http"
 "os"
 "strconv"
 "strings"
 "time"
)

type authKey string
const principalKey authKey = "principal"
type principal struct { ID int64 `json:"sub"`; Role string `json:"role"`; Exp int64 `json:"exp"` }
func issueToken(id int64,role string)(string,error){
 secret:=os.Getenv("JWT_SECRET");if len(secret)<32{return "",errors.New("JWT_SECRET must be at least 32 characters")}
 header:=base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
 payload,_:=json.Marshal(principal{ID:id,Role:role,Exp:time.Now().Add(time.Hour).Unix()})
 unsigned:=header+"."+base64.RawURLEncoding.EncodeToString(payload)
 mac:=hmac.New(sha256.New,[]byte(secret));mac.Write([]byte(unsigned))
 return unsigned+"."+base64.RawURLEncoding.EncodeToString(mac.Sum(nil)),nil
}
func parseToken(token string)(principal,error){
 var p principal
 secret:=os.Getenv("JWT_SECRET");if len(secret)<32{return p,errors.New("server JWT configuration invalid")}
 parts:=strings.Split(token,".");if len(parts)!=3{return p,errors.New("invalid token")}
 headerBytes,e:=base64.RawURLEncoding.DecodeString(parts[0]);if e!=nil{return p,e}
 var header struct{ Alg string `json:"alg"`; Typ string `json:"typ"` }
 if json.Unmarshal(headerBytes,&header)!=nil||header.Alg!="HS256"||header.Typ!="JWT"{return p,errors.New("invalid algorithm")}
 sig,e:=base64.RawURLEncoding.DecodeString(parts[2]);if e!=nil{return p,e}
 mac:=hmac.New(sha256.New,[]byte(secret));mac.Write([]byte(parts[0]+"."+parts[1]))
 if !hmac.Equal(sig,mac.Sum(nil)){return p,errors.New("invalid signature")}
 b,e:=base64.RawURLEncoding.DecodeString(parts[1]);if e!=nil{return p,e}
 if json.Unmarshal(b,&p)!=nil||p.ID<=0||p.Exp<=time.Now().Unix()||(p.Role!="member"&&p.Role!="admin"){return p,errors.New("expired or invalid claims")}
 return p,nil
}
func authError(w http.ResponseWriter,status int,code,msg string){
 w.Header().Set("Content-Type","application/json");w.WriteHeader(status)
 json.NewEncoder(w).Encode(map[string]any{"error":map[string]string{"code":code,"message":msg}})
}
func requireLogin(next http.HandlerFunc)http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  fields:=strings.Fields(r.Header.Get("Authorization"))
  if len(fields)!=2||!strings.EqualFold(fields[0],"Bearer"){authError(w,401,"UNAUTHORIZED","กรุณาเข้าสู่ระบบ");return}
  p,e:=parseToken(fields[1]);if e!=nil{authError(w,401,"UNAUTHORIZED","ตั๋วเข้าสู่ระบบไม่ถูกต้องหรือหมดอายุ");return}
  next(w,r.WithContext(context.WithValue(r.Context(),principalKey,p)))
 }
}
func requireRole(role string,next http.HandlerFunc)http.HandlerFunc{
 return requireLogin(func(w http.ResponseWriter,r *http.Request){
  p,ok:=r.Context().Value(principalKey).(principal)
  if !ok||p.Role!=role{authError(w,403,"FORBIDDEN","ไม่มีสิทธิ์เข้าถึง");return}
  next(w,r)
 })
}
func memberID(r *http.Request)int64{
 p,_:=r.Context().Value(principalKey).(principal);return p.ID
}
func pathID(r *http.Request)(int64,error){return strconv.ParseInt(r.PathValue("id"),10,64)}
