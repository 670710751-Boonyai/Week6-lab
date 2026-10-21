package main

import (
 "context"
 "encoding/json"
 "log"
 "net/http"
 "os"
 "strings"
 "time"

 "github.com/jackc/pgx/v5/pgxpool"
)

type Movie struct {
 ID int64 `json:"id"`
 Title string `json:"title"`
 Overview string `json:"overview"`
 PosterURL string `json:"posterUrl"`
}
func main() {
 url := os.Getenv("DATABASE_URL")
 if url == "" { log.Fatal("DATABASE_URL is required") }
 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
 defer cancel()
 pool, err := pgxpool.New(ctx,url)
 if err != nil { log.Fatal(err) }
 defer pool.Close()
 if err=pool.Ping(ctx);err!=nil { log.Fatal(err) }

 mux:=http.NewServeMux()
 registerRoutes(mux,pool)
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request) {
  w.Header().Set("Content-Type","application/json")
  json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
 })
 mux.HandleFunc("GET /api/movies",func(w http.ResponseWriter,r *http.Request) {
  rows,err:=pool.Query(r.Context(),`SELECT id,title,COALESCE(overview_th,overview,''),CASE WHEN poster_path IS NULL THEN '' WHEN poster_path LIKE 'https://%' THEN poster_path ELSE 'https://image.tmdb.org/t/p/w500'||poster_path END FROM movies ORDER BY popularity DESC LIMIT 100`)
  if err!=nil { http.Error(w,"database query failed",500);return }
  defer rows.Close()
  movies:=make([]Movie,0)
  for rows.Next() {
   var m Movie
   if err:=rows.Scan(&m.ID,&m.Title,&m.Overview,&m.PosterURL);err!=nil {http.Error(w,"database scan failed",500);return}
   movies=append(movies,m)
  }
  if rows.Err()!=nil {http.Error(w,"database iteration failed",500);return}
  w.Header().Set("Content-Type","application/json")
  json.NewEncoder(w).Encode(map[string]any{"items":movies})
 })
 allowed:=map[string]bool{}
 for _,origin:=range strings.Split(os.Getenv("ALLOWED_ORIGINS"),","){origin=strings.TrimSpace(origin);if origin!=""{allowed[origin]=true}}
 handler:=http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Add("Vary","Origin")
  if allowed[r.Header.Get("Origin")] {
   w.Header().Set("Access-Control-Allow-Origin",r.Header.Get("Origin"))
   w.Header().Set("Access-Control-Allow-Methods","GET, POST, PUT, DELETE, OPTIONS")
   w.Header().Set("Access-Control-Allow-Headers","Authorization, Content-Type")
  }
  if r.Method==http.MethodOptions {w.WriteHeader(http.StatusNoContent);return}
  mux.ServeHTTP(w,r)
 })
 port:=os.Getenv("PORT");if port==""{port="8080"}
 log.Printf("listening on :%s",port)
 log.Fatal(http.ListenAndServe(":"+port,handler))
}
