package mw

import (
 "context"
 "crypto/hmac"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "errors"
 "net/http"
 "os"
 "strings"
 "time"
)

type key string
const principalKey key="principal"
type Principal struct{ ID int64 `json:"sub"`; Role string `json:"role"`; Exp int64 `json:"exp"` }
func Issue(id int64,role string)(string,error){
 secret:=os.Getenv("JWT_SECRET");if len(secret)<32{return "",errors.New("JWT_SECRET must be at least 32 characters")}
 head:=base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
 payload,_:=json.Marshal(Principal{ID:id,Role:role,Exp:time.Now().Add(time.Hour).Unix()})
 data:=head+"."+base64.RawURLEncoding.EncodeToString(payload)
 mac:=hmac.New(sha256.New,[]byte(secret));mac.Write([]byte(data))
 return data+"."+base64.RawURLEncoding.EncodeToString(mac.Sum(nil)),nil
}
func verify(token string)(Principal,error){
 var p Principal
 secret:=os.Getenv("JWT_SECRET");if len(secret)<32{return p,errors.New("JWT not configured")}
 parts:=strings.Split(token,".");if len(parts)!=3{return p,errors.New("invalid token")}
 h,e:=base64.RawURLEncoding.DecodeString(parts[0]);if e!=nil{return p,e}
 var header struct{Alg string `json:"alg"`;Typ string `json:"typ"`}
 if json.Unmarshal(h,&header)!=nil||header.Alg!="HS256"||header.Typ!="JWT"{return p,errors.New("invalid header")}
 signature,e:=base64.RawURLEncoding.DecodeString(parts[2]);if e!=nil{return p,e}
 mac:=hmac.New(sha256.New,[]byte(secret));mac.Write([]byte(parts[0]+"."+parts[1]))
 if !hmac.Equal(signature,mac.Sum(nil)){return p,errors.New("bad signature")}
 payload,e:=base64.RawURLEncoding.DecodeString(parts[1]);if e!=nil{return p,e}
 if json.Unmarshal(payload,&p)!=nil||p.ID<=0||p.Exp<=time.Now().Unix()||(p.Role!="admin"&&p.Role!="member"){return p,errors.New("invalid claims")}
 return p,nil
}
func denied(w http.ResponseWriter,status int,code string){
 w.Header().Set("Content-Type","application/json");w.WriteHeader(status)
 _=json.NewEncoder(w).Encode(map[string]any{"error":map[string]string{"code":code,"message":http.StatusText(status)}})
}
func RequireLogin(next http.HandlerFunc)http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  f:=strings.Fields(r.Header.Get("Authorization"))
  if len(f)!=2||!strings.EqualFold(f[0],"Bearer"){denied(w,401,"UNAUTHORIZED");return}
  p,e:=verify(f[1]);if e!=nil{denied(w,401,"UNAUTHORIZED");return}
  next(w,r.WithContext(context.WithValue(r.Context(),principalKey,p)))
 }
}
func RequireRole(role string,next http.HandlerFunc)http.HandlerFunc{
 return func(w http.ResponseWriter,r *http.Request){
  p,ok:=r.Context().Value(principalKey).(Principal)
  if !ok||p.Role!=role{denied(w,403,"FORBIDDEN");return}
  next(w,r)
 }
}
func MemberID(r *http.Request)int64{p,_:=r.Context().Value(principalKey).(Principal);return p.ID}
