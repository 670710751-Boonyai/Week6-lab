package handler

import(
 "encoding/json"
 "errors"
 "net/http"
 "strings"
 "unicode/utf8"
 "github.com/jackc/pgx/v5/pgconn"
 "moviehub/mw"
)
func memberID(r *http.Request)int64{id,_:=mw.MemberID(r.Context());return id}
func badReference(w http.ResponseWriter,err error)bool{
 var p *pgconn.PgError
 if errors.As(err,&p)&&p.Code=="23503"{
  writeError(w,404,"NOT_FOUND","ไม่พบหนังหรือสมาชิกที่ระบุ");return true
 }
 return false
}
func(h *Handler) PostReview(w http.ResponseWriter,r *http.Request){
 movieID,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 var req struct{Text string `json:"text"`}
 if json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&req)!=nil{writeError(w,400,"BAD_JSON","อ่าน JSON ไม่สำเร็จ");return}
 req.Text=strings.TrimSpace(req.Text)
 if utf8.RuneCountInString(req.Text)<10{writeError(w,400,"BAD_REVIEW","รีวิวต้องมีอย่างน้อย 10 ตัวอักษร");return}
 saved,err:=h.Store.SaveReview(r.Context(),memberID(r),movieID,req.Text)
 if badReference(w,err){return}
 if err!=nil{internal(w,err);return}
 writeJSON(w,201,saved)
}
func(h *Handler) PutVote(w http.ResponseWriter,r *http.Request){
 movieID,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 var req struct{Score int16 `json:"score"`}
 if json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&req)!=nil{writeError(w,400,"BAD_JSON","อ่าน JSON ไม่สำเร็จ");return}
 if req.Score<1||req.Score>10{writeError(w,400,"BAD_SCORE","คะแนนต้องอยู่ระหว่าง 1 ถึง 10");return}
 err:=h.Store.SaveVote(r.Context(),memberID(r),movieID,req.Score)
 if badReference(w,err){return}
 if err!=nil{internal(w,err);return}
 writeJSON(w,200,map[string]any{"score":req.Score})
}
func(h *Handler) Wishlist(w http.ResponseWriter,r *http.Request){
 items,err:=h.Store.Wishlist(r.Context(),memberID(r));if err!=nil{internal(w,err);return}
 writeJSON(w,200,map[string]any{"items":items})
}
func(h *Handler) AddWishlist(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 err:=h.Store.AddWishlist(r.Context(),memberID(r),id)
 if badReference(w,err){return}
 if err!=nil{internal(w,err);return}
 w.WriteHeader(204)
}
func(h *Handler) RemoveWishlist(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 if err:=h.Store.RemoveWishlist(r.Context(),memberID(r),id);err!=nil{internal(w,err);return}
 w.WriteHeader(204)
}

func(h *Handler) MyMovieState(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 member:=memberID(r)
 score,err:=h.Store.GetMemberVote(r.Context(),member,id)
 if err!=nil{internal(w,err);return}
 inWishlist,err:=h.Store.InWishlist(r.Context(),member,id)
 if err!=nil{internal(w,err);return}
 writeJSON(w,200,map[string]any{"vote":score,"inWishlist":inWishlist})
}
