package main

import (
	"fmt"
	"go-git-server/app/lib"
	"log"

	git "github.com/libgit2/git2go/v34"
)

func main() {
	fmt.Println("Hello, World!")
	lib.NewStore()
	// repo, err := git.InitRepository("test", true)
	// if err != nil {
	// 	panic(err)
	// }
	// index, err := repo.Index()
	// treeId, err := index.WriteTree()
	// tree, err := repo.LookupTree(treeId)
	// message := "git repository init commit"
	// signature := &git.Signature{Email: "aaa@qq.com", Name: "owen", When: time.Now()}
	// _, err = repo.CreateCommit("refs/heads/master", signature, signature, message, tree)
	// if err != nil {
	// 	panic(err)
	// }
	repo, err := git.OpenRepository("test")

	if err != nil {
		panic(err)
	}
	defer repo.Free()
	walker, err := repo.Walk()
	if err != nil {
		log.Fatal(err)
	}
	defer walker.Free()
	err = walker.PushHead()
	if err != nil {
		log.Fatal(err)
	}

	walker.Iterate(func(commit *git.Commit) bool {
		// 处理每个提交
		fmt.Println(commit.Id(), commit.Author().Name, commit.Message(), commit.Author().Email, commit.Author().When)
		return true
	})
}
