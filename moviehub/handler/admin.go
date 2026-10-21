package handler

import (
 "encoding/json"
 "errors"
 "net/http"
 "strings"
 "time"
 "unicode/utf8"
 "moviehub/mw"
 "moviehub/store"
)

func(h *Handler) AdminStats(w http.ResponseWriter,r *http.Request){
 st,err:=h.Store.Stats(r.Context());if err!=nil{internal(w,err);return}
 writeJSON(w,200,st)
}
func(h *Handler) AdminMembers(w http.ResponseWriter,r *http.Request){
 members,err:=h.Store.ListMembers(r.Context());if err!=nil{internal(w,err);return}
 writeJSON(w,200,map[string]any{"items":members})
}
type roleReq struct{Role string `json:"role"`}
func(h *Handler) AdminSetRole(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขสมาชิกไม่ถูกต้อง");return}
 var in roleReq
 if json.NewDecoder(r.Body).Decode(&in)!=nil{writeError(w,400,"BAD_JSON","อ่าน JSON ไม่สำเร็จ");return}
 if in.Role!="member"&&in.Role!="admin"{writeError(w,400,"BAD_ROLE","role ต้องเป็น member หรือ admin");return}
 self,_:=mw.MemberID(r.Context());if self==id{writeError(w,400,"SELF_ROLE","ห้ามเปลี่ยนบทบาทของตัวเอง");return}
 m,err:=h.Store.UpdateMemberRole(r.Context(),id,in.Role)
 if errors.Is(err,store.ErrNotFound){writeError(w,404,"MEMBER_NOT_FOUND","ไม่พบสมาชิก");return}
 if err!=nil{internal(w,err);return}
 writeJSON(w,200,m)
}
func(h *Handler) Genres(w http.ResponseWriter,r *http.Request){
 genres,err:=h.Store.ListGenres(r.Context());if err!=nil{internal(w,err);return}
 writeJSON(w,200,map[string]any{"items":genres})
}
type movieReq struct{ID int64 `json:"id"`;store.MovieInput}
func trimPtr(p *string)*string{
 if p==nil{return nil}
 s:=strings.TrimSpace(*p);if s==""{return nil};return &s
}
func validateMovie(in *store.MovieInput)(string,string){
 in.Title=strings.TrimSpace(in.Title)
 in.OriginalTitle=strings.TrimSpace(in.OriginalTitle)
 in.OriginalLanguage=strings.ToLower(strings.TrimSpace(in.OriginalLanguage))
 in.TitleTh=trimPtr(in.TitleTh)
 in.Overview=trimPtr(in.Overview)
 in.OverviewTh=trimPtr(in.OverviewTh)
 in.PosterPath=trimPtr(in.PosterPath)
 in.ReleaseDate=trimPtr(in.ReleaseDate)
 if in.Title==""||utf8.RuneCountInString(in.Title)>255{return "BAD_TITLE","ต้องมีชื่อหนังและยาวไม่เกิน 255 ตัวอักษร"}
 if in.OriginalTitle==""{in.OriginalTitle=in.Title}
 if in.OriginalLanguage==""{in.OriginalLanguage="en"}
 if len(in.OriginalLanguage)!=2{return "BAD_LANGUAGE","รหัสภาษาต้องยาว 2 ตัวอักษร"}
 if in.PosterPath!=nil && len(*in.PosterPath)>500{return "BAD_POSTER","ที่อยู่รูปโปสเตอร์ต้องยาวไม่เกิน 500 ตัวอักษร"}
 if in.PosterPath!=nil && !strings.HasPrefix(*in.PosterPath,"/")&&!strings.HasPrefix(*in.PosterPath,"https://")&&!strings.HasPrefix(*in.PosterPath,"http://"){
  return "BAD_POSTER","โปสเตอร์ต้องเป็น TMDB path หรือ http(s) URL"
 }
 if in.ReleaseDate!=nil{if _,err:=time.Parse("2006-01-02",*in.ReleaseDate);err!=nil{return "BAD_DATE","วันที่ต้องเป็น YYYY-MM-DD"}}
 return "",""
}
func(h *Handler) AdminGetMovie(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 m,err:=h.Store.AdminGetMovie(r.Context(),id)
 if errors.Is(err,store.ErrNotFound){writeError(w,404,"MOVIE_NOT_FOUND","ไม่พบหนัง");return}
 if err!=nil{internal(w,err);return};writeJSON(w,200,m)
}
func(h *Handler) AdminCreateMovie(w http.ResponseWriter,r *http.Request){
 var in movieReq
 if json.NewDecoder(r.Body).Decode(&in)!=nil{writeError(w,400,"BAD_JSON","อ่าน JSON ไม่สำเร็จ");return}
 if in.ID<0{writeError(w,400,"BAD_ID","เลขหนังต้องเป็น 0 หรือค่าบวก");return}
 if code,msg:=validateMovie(&in.MovieInput);code!=""{writeError(w,400,code,msg);return}
 m,err:=h.Store.CreateMovie(r.Context(),in.ID,in.MovieInput)
 if errors.Is(err,store.ErrDuplicateID){writeError(w,409,"MOVIE_EXISTS","รหัสหนังนี้มีแล้ว");return}
 if errors.Is(err,store.ErrBadGenre){writeError(w,400,"BAD_GENRE","ไม่พบแนวหนังนี้");return}
 if err!=nil{internal(w,err);return};writeJSON(w,201,m)
}
func(h *Handler) AdminUpdateMovie(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 var in movieReq
 if json.NewDecoder(r.Body).Decode(&in)!=nil{writeError(w,400,"BAD_JSON","อ่าน JSON ไม่สำเร็จ");return}
 if code,msg:=validateMovie(&in.MovieInput);code!=""{writeError(w,400,code,msg);return}
 m,err:=h.Store.UpdateMovie(r.Context(),id,in.MovieInput)
 if errors.Is(err,store.ErrNotFound){writeError(w,404,"MOVIE_NOT_FOUND","ไม่พบหนัง");return}
 if errors.Is(err,store.ErrBadGenre){writeError(w,400,"BAD_GENRE","ไม่พบแนวหนังนี้");return}
 if err!=nil{internal(w,err);return};writeJSON(w,200,m)
}
func(h *Handler) AdminDeleteMovie(w http.ResponseWriter,r *http.Request){
 id,ok:=pathID(r,"id");if !ok{writeError(w,400,"BAD_ID","เลขหนังไม่ถูกต้อง");return}
 err:=h.Store.DeleteMovie(r.Context(),id)
 if errors.Is(err,store.ErrNotFound){writeError(w,404,"MOVIE_NOT_FOUND","ไม่พบหนัง");return}
 if err!=nil{internal(w,err);return};w.WriteHeader(http.StatusNoContent)
}
