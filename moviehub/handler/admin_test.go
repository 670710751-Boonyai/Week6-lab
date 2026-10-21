package handler

import(
 "testing"
 "moviehub/store"
)
func TestValidateMovie(t *testing.T){
 tests:=[]struct{name string;input store.MovieInput;want string}{
  {"valid",store.MovieInput{Title:"Example",OriginalLanguage:"en"},""},
  {"empty title",store.MovieInput{Title:"   "},"BAD_TITLE"},
  {"bad language",store.MovieInput{Title:"Example",OriginalLanguage:"english"},"BAD_LANGUAGE"},
  {"bad poster",store.MovieInput{Title:"Example",PosterPath:ptr("ftp://bad")},"BAD_POSTER"},
  {"bad date",store.MovieInput{Title:"Example",ReleaseDate:ptr("October 21")},"BAD_DATE"},
 }
 for _,tt:=range tests{t.Run(tt.name,func(t *testing.T){
  got,_:=validateMovie(&tt.input);if got!=tt.want{t.Fatalf("got %q want %q",got,tt.want)}
 })}
}
func ptr(s string)*string{return &s}
