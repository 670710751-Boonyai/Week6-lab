package store

import (
 "context"
 "errors"
 "time"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgconn"
)
var ErrEmailTaken = errors.New("email taken")
type Member struct{
 ID int64 `json:"id"`
 Email string `json:"email"`
 DisplayName string `json:"displayName"`
 Role string `json:"role"`
 CreatedAt time.Time `json:"createdAt"`
 PasswordHash string `json:"-"`
}
func (s *Store) CreateMember(ctx context.Context,email,hash,name string)(Member,error){
 var m Member
 err:=s.Pool.QueryRow(ctx,`INSERT INTO members(email,password_hash,display_name) VALUES($1,$2,$3) RETURNING id,email,display_name,role,created_at`,email,hash,name).Scan(&m.ID,&m.Email,&m.DisplayName,&m.Role,&m.CreatedAt)
 var p *pgconn.PgError
 if errors.As(err,&p)&&p.Code=="23505"{return Member{},ErrEmailTaken}
 return m,err
}
func(s *Store) MemberByEmail(ctx context.Context,email string)(Member,error){
 var m Member
 err:=s.Pool.QueryRow(ctx,`SELECT id,email,display_name,role,created_at,password_hash FROM members WHERE lower(email)=$1`,email).Scan(&m.ID,&m.Email,&m.DisplayName,&m.Role,&m.CreatedAt,&m.PasswordHash)
 if errors.Is(err,pgx.ErrNoRows){return Member{},ErrNotFound}
 return m,err
}
func(s *Store) MemberByID(ctx context.Context,id int64)(Member,error){
 var m Member
 err:=s.Pool.QueryRow(ctx,`SELECT id,email,display_name,role,created_at FROM members WHERE id=$1`,id).Scan(&m.ID,&m.Email,&m.DisplayName,&m.Role,&m.CreatedAt)
 if errors.Is(err,pgx.ErrNoRows){return Member{},ErrNotFound}
 return m,err
}
func(s *Store) ListMembers(ctx context.Context)([]Member,error){
 rows,err:=s.Pool.Query(ctx,`SELECT id,email,display_name,role,created_at FROM members ORDER BY created_at DESC,id DESC`)
 if err!=nil{return nil,err};defer rows.Close()
 result:=[]Member{}
 for rows.Next(){var m Member;if err:=rows.Scan(&m.ID,&m.Email,&m.DisplayName,&m.Role,&m.CreatedAt);err!=nil{return nil,err};result=append(result,m)}
 return result,rows.Err()
}
func(s *Store) UpdateMemberRole(ctx context.Context,id int64,role string)(Member,error){
 var m Member
 err:=s.Pool.QueryRow(ctx,`UPDATE members SET role=$2,updated_at=now() WHERE id=$1 RETURNING id,email,display_name,role,created_at`,id,role).Scan(&m.ID,&m.Email,&m.DisplayName,&m.Role,&m.CreatedAt)
 if errors.Is(err,pgx.ErrNoRows){return Member{},ErrNotFound}
 return m,err
}
