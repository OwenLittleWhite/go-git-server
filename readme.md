brew list libiconv
brew link --overwrite libgit2
go run -tags static,system_libgit2 main.go
