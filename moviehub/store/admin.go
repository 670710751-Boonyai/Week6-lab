package store

import (
 "context"
 "errors"
 "time"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgconn"
)
var(
 ErrDuplicateID=errors.New("duplicate movie id")
 ErrBadGenre=errors.New("invalid genre")
)
type TopMovie struct {
 ID int64 `json:"id"`
 Title string `json:"title"`
 TitleTh *string `json:"titleTh"`
 Poster *string `json:"poster"`
 VoteCount int64 `json:"voteCount"`
 VoteAverage float64 `json:"voteAverage"`
}
type Stats struct{
 Movies int64 `json:"movies"`
 Members int64 `json:"members"`
 Admins int64 `json:"admins"`
 Votes int64 `json:"votes"`
 WishlistItems int64 `json:"wishlistItems"`
 NewestMembers []Member `json:"newestMembers"`
 TopVoted []TopMovie `json:"topVoted"`
}
func(s *Store) Stats(ctx context.Context)(Stats,error){
 st:=Stats{NewestMembers:[]Member{},TopVoted:[]TopMovie{}}
 err:=s.Pool.QueryRow(ctx,`SELECT (SELECT count(*) FROM movies),(SELECT count(*) FROM members),(SELECT count(*) FROM members WHERE role='admin'),(SELECT count(*) FROM votes),(SELECT count(*) FROM wishlist_items)`).Scan(&st.Movies,&st.Members,&st.Admins,&st.Votes,&st.WishlistItems)
 if err!=nil{return st,err}
 rows,err:=s.Pool.Query(ctx,`SELECT id,email,display_name,role,created_at FROM members ORDER BY created_at DESC,id DESC LIMIT 5`)
 if err!=nil{return st,err}
 for rows.Next(){var m Member;if err=rows.Scan(&m.ID,&m.Email,&m.DisplayName,&m.Role,&m.CreatedAt);err!=nil{rows.Close();return st,err};st.NewestMembers=append(st.NewestMembers,m)}
 err=rows.Err();rows.Close();if err!=nil{return st,err}
 rows,err=s.Pool.Query(ctx,`SELECT m.id,m.title,m.title_th,`+posterExpr+`,v.vote_count,v.vote_average::float8 FROM movie_vote_stats v JOIN movies m ON m.id=v.movie_id ORDER BY v.vote_average DESC,v.vote_count DESC,m.title LIMIT 5`)
 if err!=nil{return st,err};defer rows.Close()
 for rows.Next(){var t TopMovie;if err=rows.Scan(&t.ID,&t.Title,&t.TitleTh,&t.Poster,&t.VoteCount,&t.VoteAverage);err!=nil{return st,err};st.TopVoted=append(st.TopVoted,t)}
 return st,rows.Err()
}
type Genre struct{ID int `json:"id"`; NameEn string `json:"nameEn"`;NameTh *string `json:"nameTh"`}
func(s *Store) ListGenres(ctx context.Context)([]Genre,error){
 rows,err:=s.Pool.Query(ctx,`SELECT id,name_en,name_th FROM genres ORDER BY name_en`)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]Genre{}
 for rows.Next(){var g Genre;if err:=rows.Scan(&g.ID,&g.NameEn,&g.NameTh);err!=nil{return nil,err};out=append(out,g)}
 return out,rows.Err()
}
type MovieInput struct{
 Title string `json:"title"`
 TitleTh *string `json:"titleTh"`
 OriginalTitle string `json:"originalTitle"`
 OriginalLanguage string `json:"originalLanguage"`
 Overview *string `json:"overview"`
 OverviewTh *string `json:"overviewTh"`
 PosterPath *string `json:"posterPath"`
 ReleaseDate *string `json:"releaseDate"`
 GenreID *int `json:"genreId"`
}
type Impact struct{Reviews int64 `json:"reviews"`;Votes int64 `json:"votes"`;Wishlist int64 `json:"wishlist"`}
type AdminMovie struct{
 ID int64 `json:"id"`
 MovieInput
 Poster *string `json:"poster"`
 VoteAverage float64 `json:"voteAverage"`
 UpdatedAt time.Time `json:"updatedAt"`
 Impact Impact `json:"impact"`
}
func(s *Store) AdminGetMovie(ctx context.Context,id int64)(AdminMovie,error){
 var m AdminMovie;var release *time.Time
 err:=s.Pool.QueryRow(ctx,`SELECT m.id,m.title,m.title_th,m.original_title,m.original_language,m.overview,m.overview_th,m.poster_path,m.release_date,
 (SELECT mg.genre_id FROM movie_genres mg WHERE mg.movie_id=m.id AND mg.position=0),
 `+posterExpr+`,round(m.vote_average,1)::float8,m.updated_at,
 (SELECT count(*) FROM reviews WHERE movie_id=m.id),(SELECT count(*) FROM votes WHERE movie_id=m.id),
 (SELECT count(*) FROM wishlist_items WHERE movie_id=m.id)
 FROM movies m WHERE m.id=$1`,id).
 Scan(&m.ID,&m.Title,&m.TitleTh,&m.OriginalTitle,&m.OriginalLanguage,&m.Overview,&m.OverviewTh,&m.PosterPath,&release,&m.GenreID,&m.Poster,&m.VoteAverage,&m.UpdatedAt,&m.Impact.Reviews,&m.Impact.Votes,&m.Impact.Wishlist)
 if errors.Is(err,pgx.ErrNoRows){return m,ErrNotFound}
 if release!=nil{d:=release.Format("2006-01-02");m.ReleaseDate=&d}
 return m,err
}
func setMainGenre(ctx context.Context,tx pgx.Tx,id int64,genre *int)error{
 if genre==nil{_,err:=tx.Exec(ctx,`DELETE FROM movie_genres WHERE movie_id=$1 AND position=0`,id);return err}
 // Validate before mutating to give a readable BAD_GENRE response.
 var exists bool
 if err:=tx.QueryRow(ctx,`SELECT EXISTS(SELECT 1 FROM genres WHERE id=$1)`,*genre).Scan(&exists);err!=nil{return err}
 if !exists{return ErrBadGenre}
 if _,err:=tx.Exec(ctx,`DELETE FROM movie_genres WHERE movie_id=$1 AND (position=0 OR genre_id=$2)`,id,*genre);err!=nil{return err}
 _,err:=tx.Exec(ctx,`INSERT INTO movie_genres(movie_id,genre_id,position) VALUES($1,$2,0)`,id,*genre)
 return err
}
func(s *Store) CreateMovie(ctx context.Context,id int64,in MovieInput)(AdminMovie,error){
 tx,err:=s.Pool.Begin(ctx);if err!=nil{return AdminMovie{},err};defer tx.Rollback(ctx)
 if id==0 {
  if _,err=tx.Exec(ctx,`SELECT pg_advisory_xact_lock(802412)`);err!=nil{return AdminMovie{},err}
  err=tx.QueryRow(ctx,`SELECT GREATEST(COALESCE(MAX(id),0)+1,9000000) FROM movies`).Scan(&id)
  if err!=nil{return AdminMovie{},err}
 }
 _,err=tx.Exec(ctx,`INSERT INTO movies(id,title,title_th,original_title,original_language,overview,overview_th,poster_path,release_date)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9::date)`,id,in.Title,in.TitleTh,in.OriginalTitle,in.OriginalLanguage,in.Overview,in.OverviewTh,in.PosterPath,in.ReleaseDate)
 var pe *pgconn.PgError
 if errors.As(err,&pe)&&pe.Code=="23505"{return AdminMovie{},ErrDuplicateID}
 if err!=nil{return AdminMovie{},err}
 if err=setMainGenre(ctx,tx,id,in.GenreID);err!=nil{return AdminMovie{},err}
 if err=tx.Commit(ctx);err!=nil{return AdminMovie{},err}
 return s.AdminGetMovie(ctx,id)
}
func(s *Store) UpdateMovie(ctx context.Context,id int64,in MovieInput)(AdminMovie,error){
 tx,err:=s.Pool.Begin(ctx);if err!=nil{return AdminMovie{},err};defer tx.Rollback(ctx)
 tag,err:=tx.Exec(ctx,`UPDATE movies SET title=$2,title_th=$3,original_title=$4,original_language=$5,overview=$6,overview_th=$7,poster_path=$8,release_date=$9::date,updated_at=now() WHERE id=$1`,id,in.Title,in.TitleTh,in.OriginalTitle,in.OriginalLanguage,in.Overview,in.OverviewTh,in.PosterPath,in.ReleaseDate)
 if err!=nil{return AdminMovie{},err}
 if tag.RowsAffected()==0{return AdminMovie{},ErrNotFound}
 if err=setMainGenre(ctx,tx,id,in.GenreID);err!=nil{return AdminMovie{},err}
 if err=tx.Commit(ctx);err!=nil{return AdminMovie{},err}
 return s.AdminGetMovie(ctx,id)
}
func(s *Store) DeleteMovie(ctx context.Context,id int64)error{
 tag,err:=s.Pool.Exec(ctx,`DELETE FROM movies WHERE id=$1`,id)
 if err!=nil{return err};if tag.RowsAffected()==0{return ErrNotFound};return nil
}
