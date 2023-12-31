package lib

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	_path "path/filepath"
	"strings"
	"time"

	git "github.com/libgit2/git2go/v34"
)

type Store struct {
	author    *git.Signature
	storePath string
}

type FileParams struct {
	repoPath string
	filePath string
	content  string
	message  string
	ref      string
	author   *git.Signature
	commitId string
	// encoding string
}

func NewStore() *Store {
	return &Store{
		author:    &git.Signature{Name: "bot", Email: "bot@tatfook-git.com"},
		storePath: "/Users/owenzhao/workspace/go-git-server/data",
	}
}

func (s *Store) OpenRepository(repoPath string) (repo *git.Repository, err error) {
	encodeRepoPath := base64.StdEncoding.EncodeToString([]byte(repoPath))
	path := _path.Join(s.storePath, encodeRepoPath)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating directory: %v", err)
		}
		repo, err = git.InitRepository(path, true)
		if err != nil {
			return nil, fmt.Errorf("error creating repository: %v", err)
		}
		index, err := repo.Index()
		if err != nil {
			return nil, fmt.Errorf("error creating repo index: %v", err)
		}
		treeId, err := index.WriteTree()
		if err != nil {
			return nil, err
		}
		tree, err := repo.LookupTree(treeId)
		if err != nil {
			return nil, err
		}
		message := "git repository init commit"
		_, err = repo.CreateCommit("refs/heads/master", s.author, s.author, message, tree)
		if err != nil {
			return nil, err
		}
	} else {
		repo, err = git.OpenRepository(path)
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}
func (s *Store) SaveFile(p *FileParams) (commitId string, err error) {
	if p.repoPath == "" || p.filePath == "" {
		return "", fmt.Errorf("should have repoPath and filePath params")
	}
	p.ref = formatRef(p.ref, p.filePath)
	if p.message == "" {
		p.message = fmt.Sprintf("save file %s", p.filePath)
	}
	if p.author == nil {
		p.author = s.author
		p.author.When = time.Now()
	}
	repo, err := s.OpenRepository(p.repoPath)

	defer repo.Free()

	if err != nil {
		return
	}
	ref, err := repo.References.Lookup(p.ref)
	defer ref.Free()
	if err != nil {
		return
	}
	refCommit, err := repo.LookupCommit(ref.Target())
	defer refCommit.Free()
	if err != nil {
		return
	}
	fileId, err := repo.CreateBlobFromBuffer([]byte(p.content))
	if err != nil {
		return
	}

	treeBuilder, err := repo.TreeBuilder()
	defer treeBuilder.Free()
	if err != nil {
		return
	}
	treeBuilder.Insert(p.filePath, fileId, git.FilemodeBlob)
	treeId, err := treeBuilder.Write()
	if err != nil {
		return
	}
	tree, err := repo.LookupTree(treeId)
	defer tree.Free()
	if err != nil {
		log.Fatal(err)
	}
	commitOId, err := repo.CreateCommit(p.ref, p.author, p.author, p.message, tree, refCommit)
	if err != nil {
		return
	}

	return commitOId.String(), nil
}

func (s *Store) DeleteFile(p *FileParams) (commitId string, err error) {
	if p.repoPath == "" || p.filePath == "" {
		return "", fmt.Errorf("should have repoPath and filePath params")
	}
	p.ref = formatRef(p.ref, p.filePath)
	if p.message == "" {
		p.message = fmt.Sprintf("delete file %s", p.filePath)
	}
	if p.author == nil {
		p.author = s.author
		p.author.When = time.Now()
	}
	repo, err := s.OpenRepository(p.repoPath)
	defer repo.Free()
	if err != nil {
		return
	}
	ref, err := repo.References.Lookup(p.ref)

	if err != nil {
		return
	}
	refCommit, err := repo.LookupCommit(ref.Target())
	if err != nil {
		return
	}
	latestTree, err := refCommit.Tree()
	treeBuilder, err := repo.TreeBuilderFromTree(latestTree)
	err = treeBuilder.Remove(p.filePath)
	if err != nil {
		fmt.Printf("Err when remove: %v\n", err)
		return "", nil
	}
	oId, err := treeBuilder.Write()
	if err != nil {
		return
	}
	tree, err := repo.LookupTree(oId)
	if err != nil {
		return
	}
	commitOId, err := repo.CreateCommit(p.ref, p.author, p.author, p.message, tree, refCommit)
	commitId = commitOId.String()
	return
}

func (s *Store) GetBlob(p *FileParams) (content string, commitId string, err error) {
	if p.repoPath == "" || p.filePath == "" {
		return "", "", fmt.Errorf("should have repoPath and filePath params")
	}
	p.ref = formatRef(p.ref, p.filePath)
	repo, err := s.OpenRepository(p.repoPath)
	defer repo.Free()
	var tree *git.Tree
	if p.commitId != "" {
		oId, err := git.NewOid(p.commitId)
		if err != nil {
			return "", "", fmt.Errorf("provided commitId error: %v", err)
		}
		commit, err := repo.LookupCommit(oId)
		if err != nil {
			return "", "", err
		}
		tree, err = commit.Tree()
		if err != nil {
			return "", "", err
		}
	} else {
		ref, err := repo.References.Lookup(p.ref)
		if err != nil {
			return "", "", err
		}
		oId := ref.Target()
		commitId = oId.String()
		fmt.Printf("commitId: %v\n", commitId)
		commit, err := repo.LookupCommit(oId)
		if err != nil {
			return "", "", err
		}
		tree, err = commit.Tree()
		fmt.Printf("tree: %v\n", tree)
		if err != nil {
			return "", "", err
		}
	}
	entry := tree.EntryByName(p.filePath)
	fmt.Printf("entry: %v\n", entry)
	blob, err := repo.LookupBlob(entry.Id)
	content = string(blob.Contents())
	return
}

func formatRef(ref string, filePath string) string {
	if len(ref) > 0 && strings.Index(ref, "refs/heads") == 0 {
		return ref
	}
	if len(ref) > 0 {
		if ref == "master" {
			return _path.Join("refs", "heads", "master")
		}
		return _path.Join("refs", "heads", "self", ref)
	}
	// return _path.Join("refs", "heads", "sys", base64.StdEncoding.EncodeToString([]byte(filePath)))
	return _path.Join("refs", "heads", "master")
}
