package lib

import (
	"fmt"
	"testing"
)

func TestOpenRepository(t *testing.T) {
	store := NewStore()
	result, err := store.OpenRepository("test")
	if err != nil {
		t.Fail()
	}
	fmt.Printf("OpenRepository, %v", result)
}

func TestSaveFile(t *testing.T) {
	store := NewStore()
	result, err := store.SaveFile(&FileParams{repoPath: "test", filePath: "hello.txt", content: "hello world催催催!!!"})
	if err != nil {
		fmt.Printf("Err, %s\n", err)
		t.Fail()
	}
	fmt.Printf("SaveFile, %s", result)
}

func TestDeleteFile(t *testing.T) {
	store := NewStore()
	result, err := store.DeleteFile(&FileParams{repoPath: "test", filePath: "hello3.txt"})
	if err != nil {
		fmt.Printf("Err, %s\n", err)
		t.Fail()
	}
	fmt.Printf("Delete, %s", result)
}

func TestGetBlob(t *testing.T) {
	store := NewStore()
	content, commitId, err := store.GetBlob(&FileParams{repoPath: "test", filePath: "hello.txt"})
	if err != nil {
		fmt.Printf("Err, %s\n", err)
		t.Fail()
	}
	fmt.Printf("content:%s, commitId: %s\n", content, commitId)

	content, commitId, err = store.GetBlob(&FileParams{repoPath: "test", filePath: "hello.txt", commitId: commitId})
	if err != nil {
		fmt.Printf("Err from commitId, %s\n", err)
		t.Fail()
	}
	fmt.Printf("fetch from commitId, content:%s, commitId: %s", content, commitId)
}
