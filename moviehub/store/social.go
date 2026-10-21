package store

import (
 "context"
 "errors"
 "github.com/jackc/pgx/v5"
)
func(s *Store) SaveReview(ctx context.Context,memberID,movieID int64,body string)(Review,error){
 var id int64
 err:=s.Pool.QueryRow(ctx,`INSERT INTO reviews(member_id,movie_id,text) VALUES($1,$2,$3)
 ON CONFLICT(member_id,movie_id) DO UPDATE SET text=EXCLUDED.text,updated_at=now()
 RETURNING id`,memberID,movieID,body).Scan(&id)
 if err!=nil{return Review{},err}
 rows,err:=s.ListReviews(ctx,movieID);if err!=nil{return Review{},err}
 for _,r:=range rows{if r.ID==id{return r,nil}}
 return Review{},ErrNotFound
}
func(s *Store) SaveVote(ctx context.Context,memberID,movieID int64,score int16)error{
 _,err:=s.Pool.Exec(ctx,`INSERT INTO votes(member_id,movie_id,score) VALUES($1,$2,$3)
 ON CONFLICT(member_id,movie_id) DO UPDATE SET score=EXCLUDED.score,updated_at=now()`,memberID,movieID,score)
 return err
}
func(s *Store) Wishlist(ctx context.Context,memberID int64)([]Movie,error){
 rows,err:=s.Pool.Query(ctx,movieSelect+` JOIN wishlist_items wi ON wi.movie_id=m.id WHERE wi.member_id=$1 ORDER BY wi.added_at DESC`,memberID)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]Movie{}
 for rows.Next(){m,err:=scanMovie(rows);if err!=nil{return nil,err};out=append(out,m)}
 return out,rows.Err()
}
func(s *Store) AddWishlist(ctx context.Context,memberID,movieID int64)error{
 _,err:=s.Pool.Exec(ctx,`INSERT INTO wishlist_items(member_id,movie_id) VALUES($1,$2) ON CONFLICT(member_id,movie_id) DO NOTHING`,memberID,movieID)
 return err
}
func(s *Store) RemoveWishlist(ctx context.Context,memberID,movieID int64)error{
 _,err:=s.Pool.Exec(ctx,`DELETE FROM wishlist_items WHERE member_id=$1 AND movie_id=$2`,memberID,movieID)
 return err
}
func(s *Store) GetMemberVote(ctx context.Context,memberID,movieID int64)(*int16,error){
 var score int16
 err:=s.Pool.QueryRow(ctx,`SELECT score FROM votes WHERE member_id=$1 AND movie_id=$2`,memberID,movieID).Scan(&score)
 if errors.Is(err,pgx.ErrNoRows){return nil,nil}
 if err!=nil{return nil,err};return &score,nil
}
func(s *Store) InWishlist(ctx context.Context,memberID,movieID int64)(bool,error){
 var ok bool
 err:=s.Pool.QueryRow(ctx,`SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE member_id=$1 AND movie_id=$2)`,memberID,movieID).Scan(&ok)
 return ok,err
}
