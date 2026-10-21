package main

import(
 "context"
 "log"
 "net/http"
 "os"
 "strings"
 "time"
 "github.com/jackc/pgx/v5/pgxpool"
 "moviehub/auth"
 "moviehub/handler"
 "moviehub/mw"
 "moviehub/store"
)
func main(){
 secret:=os.Getenv("JWT_SECRET")
 if len(secret)<32{log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")}
 pool,err:=pgxpool.New(context.Background(),os.Getenv("DATABASE_URL"))
 if err!=nil{log.Fatal(err)};defer pool.Close()
 if err=pool.Ping(context.Background());err!=nil{log.Fatal("ต่อฐานข้อมูลไม่ติด: ",err)}
 tickets:=&auth.Issuer{Secret:[]byte(secret),TTL:time.Hour}
 h:=&handler.Handler{Store:&store.Store{Pool:pool},Tickets:tickets}
 login:=func(next http.HandlerFunc)http.HandlerFunc{return mw.RequireLogin(tickets,next)}
 admin:=func(next http.HandlerFunc)http.HandlerFunc{return login(mw.RequireRole("admin",next))}
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /health",func(w http.ResponseWriter,r *http.Request){w.WriteHeader(200);_,_=w.Write([]byte("ok"))})
 mux.HandleFunc("POST /api/auth/register",h.Register)
 mux.HandleFunc("POST /api/auth/login",h.Login)
 mux.HandleFunc("GET /api/me",login(h.Me))
 mux.HandleFunc("GET /api/genres",h.Genres)
 mux.HandleFunc("GET /api/movies",h.SearchMovies)
 mux.HandleFunc("GET /api/movies/{id}",h.GetMovie)
 mux.HandleFunc("GET /api/movies/{id}/reviews",h.ListReviews)
 mux.HandleFunc("POST /api/movies/{id}/reviews",login(h.PostReview))
 mux.HandleFunc("PUT /api/movies/{id}/vote",login(h.PutVote))
 mux.HandleFunc("GET /api/movies/{id}/my-state",login(h.MyMovieState))
 mux.HandleFunc("GET /api/me/wishlist",login(h.Wishlist))
 mux.HandleFunc("PUT /api/me/wishlist/{id}",login(h.AddWishlist))
 mux.HandleFunc("DELETE /api/me/wishlist/{id}",login(h.RemoveWishlist))
 mux.HandleFunc("GET /api/admin/stats",admin(h.AdminStats))
 mux.HandleFunc("GET /api/admin/members",admin(h.AdminMembers))
 mux.HandleFunc("PUT /api/admin/members/{id}/role",admin(h.AdminSetRole))
 mux.HandleFunc("GET /api/admin/movies/{id}",admin(h.AdminGetMovie))
 mux.HandleFunc("POST /api/admin/movies",admin(h.AdminCreateMovie))
 mux.HandleFunc("PUT /api/admin/movies/{id}",admin(h.AdminUpdateMovie))
 mux.HandleFunc("DELETE /api/admin/movies/{id}",admin(h.AdminDeleteMovie))
 port:=os.Getenv("PORT");if port==""{port="8080"}
 origins:=strings.Split(os.Getenv("ALLOWED_ORIGINS"),",")
 log.Printf("MovieHub API listening on :%s",port)
 log.Fatal(http.ListenAndServe(":"+port,mw.Logging(mw.CORS(origins,mux))))
}
