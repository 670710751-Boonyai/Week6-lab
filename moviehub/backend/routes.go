package main

import (
 "encoding/json"
 "errors"
 "net/http"
 "strconv"
 "strings"
 "time"
 "fmt"

 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
 "golang.org/x/crypto/bcrypt"
)

func send(w http.ResponseWriter,status int,data any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(data)}
func fail(w http.ResponseWriter,status int,msg string){authError(w,status,http.StatusText(status),msg)}
func decode(r *http.Request,v any)error{r.Body=http.MaxBytesReader(nil,r.Body,1<<20);d:=json.NewDecoder(r.Body);d.DisallowUnknownFields();return d.Decode(v)}
func idOf(r *http.Request)int64{id,_:=strconv.ParseInt(r.PathValue("id"),10,64);return id}
func registerRoutes(mux *http.ServeMux,db *pgxpool.Pool){
 mux.HandleFunc("POST /api/auth/register",func(w http.ResponseWriter,r *http.Request){
  var x struct{Email string `json:"email"`;Password string `json:"password"`;DisplayName string `json:"displayName"`}
  if decode(r,&x)!=nil {fail(w,400,"JSON ไม่ถูกต้อง");return}
  x.Email=strings.ToLower(strings.TrimSpace(x.Email));x.DisplayName=strings.TrimSpace(x.DisplayName)
  if !strings.Contains(x.Email,"@")||len(x.Password)<8||len(x.DisplayName)==0||len(x.DisplayName)>50 {fail(w,400,"ข้อมูลสมัครสมาชิกไม่ถูกต้อง");return}
  hash,e:=bcrypt.GenerateFromPassword([]byte(x.Password),bcrypt.DefaultCost);if e!=nil{fail(w,500,"hash failed");return}
  var id int64
  e=db.QueryRow(r.Context(),"INSERT INTO members(email,password_hash,display_name) VALUES($1,$2,$3) RETURNING id",x.Email,string(hash),x.DisplayName).Scan(&id)
  if e!=nil{fail(w,409,"อีเมลถูกใช้แล้วหรือข้อมูลไม่ถูกต้อง");return}
  send(w,201,map[string]any{"id":id})
 })
 mux.HandleFunc("POST /api/auth/login",func(w http.ResponseWriter,r *http.Request){
  var x struct{Email string `json:"email"`;Password string `json:"password"`}
  if decode(r,&x)!=nil{fail(w,400,"JSON ไม่ถูกต้อง");return}
  var id int64;var hash,name,role string
  e:=db.QueryRow(r.Context(),"SELECT id,password_hash,display_name,role FROM members WHERE lower(email)=lower($1)",strings.TrimSpace(x.Email)).Scan(&id,&hash,&name,&role)
  if e!=nil||bcrypt.CompareHashAndPassword([]byte(hash),[]byte(x.Password))!=nil{fail(w,401,"อีเมลหรือรหัสผ่านไม่ถูกต้อง");return}
  token,e:=issueToken(id,role);if e!=nil{fail(w,500,"JWT configuration error");return}
  send(w,200,map[string]any{"token":token,"member":map[string]any{"id":id,"email":strings.ToLower(x.Email),"displayName":name,"role":role}})
 })
 mux.HandleFunc("GET /api/me",requireLogin(func(w http.ResponseWriter,r *http.Request){
  var email,name,role string
  e:=db.QueryRow(r.Context(),"SELECT email,display_name,role FROM members WHERE id=$1",memberID(r)).Scan(&email,&name,&role)
  if e!=nil{fail(w,404,"ไม่พบสมาชิก");return}
  send(w,200,map[string]any{"id":memberID(r),"email":email,"displayName":name,"role":role})
 }))
 mux.HandleFunc("GET /api/movies/{id}",func(w http.ResponseWriter,r *http.Request){
  id:=idOf(r);if id<=0{fail(w,400,"เลขหนังไม่ถูกต้อง");return}
  var m Movie
  e:=db.QueryRow(r.Context(),`SELECT id,title,COALESCE(overview_th,overview,''),COALESCE('https://image.tmdb.org/t/p/w500'||poster_path,'') FROM movies WHERE id=$1`,id).Scan(&m.ID,&m.Title,&m.Overview,&m.PosterURL)
  if e!=nil{fail(w,404,"ไม่พบหนัง");return}
  send(w,200,map[string]any{"movie":m})
 })
 mux.HandleFunc("GET /api/movies/{id}/reviews",func(w http.ResponseWriter,r *http.Request){
  rows,e:=db.Query(r.Context(),`SELECT r.id,r.text,m.display_name,r.created_at FROM reviews r JOIN members m ON m.id=r.member_id WHERE r.movie_id=$1 ORDER BY r.created_at DESC`,idOf(r))
  if e!=nil{fail(w,500,"อ่านรีวิวไม่สำเร็จ");return};defer rows.Close()
  items:=[]any{}
  for rows.Next(){var id int64;var text,name string;var created time.Time;if rows.Scan(&id,&text,&name,&created)!=nil{fail(w,500,"อ่านรีวิวไม่สำเร็จ");return};items=append(items,map[string]any{"id":id,"text":text,"displayName":name,"createdAt":created})}
  if rows.Err()!=nil{fail(w,500,"อ่านรีวิวไม่สำเร็จ");return}
  send(w,200,map[string]any{"items":items})
 })
 mux.HandleFunc("POST /api/movies/{id}/reviews",requireLogin(func(w http.ResponseWriter,r *http.Request){
  var x struct{Text string `json:"text"`};if decode(r,&x)!=nil||len([]rune(strings.TrimSpace(x.Text)))<10{fail(w,400,"รีวิวต้องมีอย่างน้อย 10 ตัวอักษร");return}
  _,e:=db.Exec(r.Context(),`INSERT INTO reviews(member_id,movie_id,text) VALUES($1,$2,$3) ON CONFLICT(member_id,movie_id) DO UPDATE SET text=excluded.text,updated_at=now()`,memberID(r),idOf(r),x.Text)
  if e!=nil{fail(w,400,"บันทึกรีวิวไม่สำเร็จ");return};send(w,200,map[string]bool{"ok":true})
 }))
 mux.HandleFunc("PUT /api/movies/{id}/vote",requireLogin(func(w http.ResponseWriter,r *http.Request){
  var x struct{Score int `json:"score"`};if decode(r,&x)!=nil||x.Score<1||x.Score>10{fail(w,400,"คะแนนต้องอยู่ระหว่าง 1–10");return}
  _,e:=db.Exec(r.Context(),`INSERT INTO votes(member_id,movie_id,score) VALUES($1,$2,$3) ON CONFLICT(member_id,movie_id) DO UPDATE SET score=excluded.score,updated_at=now()`,memberID(r),idOf(r),x.Score)
  if e!=nil{fail(w,400,"บันทึกคะแนนไม่สำเร็จ");return};send(w,200,map[string]bool{"ok":true})
 }))
 mux.HandleFunc("GET /api/me/wishlist",requireLogin(func(w http.ResponseWriter,r *http.Request){
  rows,e:=db.Query(r.Context(),`SELECT m.id,m.title,COALESCE(m.overview_th,m.overview,''),COALESCE('https://image.tmdb.org/t/p/w500'||m.poster_path,'') FROM wishlist_items wi JOIN movies m ON m.id=wi.movie_id WHERE wi.member_id=$1 ORDER BY wi.added_at DESC`,memberID(r))
  if e!=nil{fail(w,500,"อ่าน wishlist ไม่สำเร็จ");return};defer rows.Close()
  items:=[]Movie{};for rows.Next(){var m Movie;if rows.Scan(&m.ID,&m.Title,&m.Overview,&m.PosterURL)!=nil{fail(w,500,"อ่าน wishlist ไม่สำเร็จ");return};items=append(items,m)}
  if rows.Err()!=nil{fail(w,500,"อ่าน wishlist ไม่สำเร็จ");return};send(w,200,map[string]any{"items":items})
 }))
 mux.HandleFunc("PUT /api/me/wishlist/{id}",requireLogin(func(w http.ResponseWriter,r *http.Request){
  _,e:=db.Exec(r.Context(),"INSERT INTO wishlist_items(member_id,movie_id) VALUES($1,$2) ON CONFLICT DO NOTHING",memberID(r),idOf(r))
  if e!=nil{fail(w,400,"เพิ่ม wishlist ไม่สำเร็จ");return};w.WriteHeader(204)
 }))
 mux.HandleFunc("DELETE /api/me/wishlist/{id}",requireLogin(func(w http.ResponseWriter,r *http.Request){
  _,e:=db.Exec(r.Context(),"DELETE FROM wishlist_items WHERE member_id=$1 AND movie_id=$2",memberID(r),idOf(r))
  if e!=nil{fail(w,500,"ลบ wishlist ไม่สำเร็จ");return};w.WriteHeader(204)
 }))
 mux.HandleFunc("GET /api/admin/stats",requireRole("admin",func(w http.ResponseWriter,r *http.Request){
  counts:=map[string]int64{}
  for _,table:=range []string{"movies","members","reviews","votes","wishlist_items"}{var n int64;e:=db.QueryRow(r.Context(),"SELECT count(*) FROM "+table).Scan(&n);if e!=nil{fail(w,500,"อ่านสถิติไม่สำเร็จ");return};counts[table]=n}
  send(w,200,counts)
 }))
 mux.HandleFunc("GET /api/admin/members",requireRole("admin",func(w http.ResponseWriter,r *http.Request){
  rows,e:=db.Query(r.Context(),"SELECT id,email,display_name,role FROM members ORDER BY id")
  if e!=nil{fail(w,500,"อ่านสมาชิกไม่สำเร็จ");return};defer rows.Close()
  items:=[]any{};for rows.Next(){var id int64;var email,name,role string;if rows.Scan(&id,&email,&name,&role)!=nil{fail(w,500,"อ่านสมาชิกไม่สำเร็จ");return};items=append(items,map[string]any{"id":id,"email":email,"displayName":name,"role":role})}
  if rows.Err()!=nil{fail(w,500,"อ่านสมาชิกไม่สำเร็จ");return};send(w,200,map[string]any{"items":items})
 }))
 mux.HandleFunc("PUT /api/admin/members/{id}/role",requireRole("admin",func(w http.ResponseWriter,r *http.Request){
  id:=idOf(r);var x struct{Role string `json:"role"`}
  if id<=0||decode(r,&x)!=nil||(x.Role!="admin"&&x.Role!="member"){fail(w,400,"บทบาทไม่ถูกต้อง");return}
  if id==memberID(r){fail(w,400,"ห้ามเปลี่ยนบทบาทตัวเอง");return}
  tag,e:=db.Exec(r.Context(),"UPDATE members SET role=$1,updated_at=now() WHERE id=$2",x.Role,id)
  if e!=nil{fail(w,500,"แก้ไขบทบาทไม่สำเร็จ");return};if tag.RowsAffected()==0{fail(w,404,"ไม่พบสมาชิก");return}
  send(w,200,map[string]any{"id":id,"role":x.Role})
 }))
 mux.HandleFunc("GET /api/admin/genres",requireRole("admin",func(w http.ResponseWriter,r *http.Request){
  rows,e:=db.Query(r.Context(),"SELECT id,name_en,COALESCE(name_th,'') FROM genres ORDER BY name_en");if e!=nil{fail(w,500,"อ่านประเภทหนังไม่สำเร็จ");return};defer rows.Close()
  items:=[]any{};for rows.Next(){var id int;var en,th string;if rows.Scan(&id,&en,&th)!=nil{fail(w,500,"อ่านประเภทหนังไม่สำเร็จ");return};items=append(items,map[string]any{"id":id,"nameEn":en,"nameTh":th})}
  if rows.Err()!=nil{fail(w,500,"อ่านประเภทหนังไม่สำเร็จ");return};send(w,200,map[string]any{"items":items})
 }))
 mux.HandleFunc("POST /api/admin/movies",requireRole("admin",func(w http.ResponseWriter,r *http.Request){saveMovie(w,r,db,false)}))
 mux.HandleFunc("PUT /api/admin/movies/{id}",requireRole("admin",func(w http.ResponseWriter,r *http.Request){saveMovie(w,r,db,true)}))
 mux.HandleFunc("DELETE /api/admin/movies/{id}",requireRole("admin",func(w http.ResponseWriter,r *http.Request){
  tag,e:=db.Exec(r.Context(),"DELETE FROM movies WHERE id=$1",idOf(r));if e!=nil{fail(w,500,"ลบหนังไม่สำเร็จ");return};if tag.RowsAffected()==0{fail(w,404,"ไม่พบหนัง");return};w.WriteHeader(204)
 }))
}
type movieInput struct{ID int64 `json:"id"`;Title string `json:"title"`;OriginalTitle string `json:"originalTitle"`;OriginalLanguage string `json:"originalLanguage"`;Overview string `json:"overview"`;PosterPath string `json:"posterPath"`}
func saveMovie(w http.ResponseWriter,r *http.Request,db *pgxpool.Pool,update bool){
 var m movieInput
 if decode(r,&m)!=nil{fail(w,400,"JSON ไม่ถูกต้อง");return}
 m.Title=strings.TrimSpace(m.Title)
 if m.Title==""||len(m.Title)>255{fail(w,400,"ชื่อหนังไม่ถูกต้อง");return}
 if m.OriginalTitle==""{m.OriginalTitle=m.Title};if m.OriginalLanguage==""{m.OriginalLanguage="en"}
 if len(m.OriginalLanguage)!=2{fail(w,400,"รหัสภาษาต้องมี 2 ตัวอักษร");return}
 var id int64;var e error
 if update{
  id=idOf(r);if id<=0{fail(w,400,"เลขหนังไม่ถูกต้อง");return}
  tag,err:=db.Exec(r.Context(),`UPDATE movies SET title=$1,original_title=$2,original_language=$3,overview=$4,poster_path=NULLIF($5,''),updated_at=now() WHERE id=$6`,m.Title,m.OriginalTitle,m.OriginalLanguage,m.Overview,m.PosterPath,id)
  e=err;if e==nil&&tag.RowsAffected()==0{fail(w,404,"ไม่พบหนัง");return}
 }else{
  id=m.ID;if id<=0{fail(w,400,"ต้องระบุ TMDB movie ID");return}
  e=db.QueryRow(r.Context(),`INSERT INTO movies(id,title,original_title,original_language,overview,poster_path) VALUES($1,$2,$3,$4,$5,NULLIF($6,'')) RETURNING id`,id,m.Title,m.OriginalTitle,m.OriginalLanguage,m.Overview,m.PosterPath).Scan(&id)
 }
 if e!=nil{if errors.Is(e,pgx.ErrNoRows){fail(w,404,"ไม่พบหนัง")}else{fail(w,400,fmt.Sprintf("บันทึกหนังไม่สำเร็จ: %v",e))};return}
 send(w,200,map[string]any{"id":id,"title":m.Title})
}
