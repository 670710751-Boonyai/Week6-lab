package store

import (
 "context"
 "github.com/jackc/pgx/v5/pgxpool"
)
type AdminStore struct{Pool *pgxpool.Pool}
type Stats struct{Movies int64 `json:"movies"`;Members int64 `json:"members"`;Reviews int64 `json:"reviews"`;Votes int64 `json:"votes"`;Wishlist int64 `json:"wishlist"`}
func(s AdminStore)Stats(ctx context.Context)(Stats,error){
 var v Stats
 for _,q:=range []struct{sql string;dest *int64}{
 {"SELECT count(*) FROM movies",&v.Movies},{"SELECT count(*) FROM members",&v.Members},{"SELECT count(*) FROM reviews",&v.Reviews},{"SELECT count(*) FROM votes",&v.Votes},{"SELECT count(*) FROM wishlist_items",&v.Wishlist},
 }{if e:=s.Pool.QueryRow(ctx,q.sql).Scan(q.dest);e!=nil{return Stats{},e}}
 return v,nil
}
