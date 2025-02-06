package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v69/github"
	"k8s.io/client-go/kubernetes"
)

type server struct {
	client           *kubernetes.Clientset
	githubClient     *github.Client
	webhookSecretKey string
}

func (s server) webhook(w http.ResponseWriter, req *http.Request) {
	payload, err := github.ValidatePayload(req, []byte(s.webhookSecretKey))
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("error: %s\n", err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("error: %s\n", err)
		return
	}

	switch event := event.(type) {
	case *github.Hook:
		fmt.Printf("Received hook event: %s\n", event)
	case *github.PushEvent:
		ctx := context.Background()
		files := getFiles(event.Commits)
		fmt.Printf("Received push event to %s\n", *event.Repo.FullName)
		fmt.Printf("Files changed: %v\n", strings.Join(files, ", "))
		for _, file := range files {
			downloadedFile, _, err := s.githubClient.Repositories.DownloadContents(
				ctx, *event.Repo.Owner.Name, *event.Repo.Name, file, &github.RepositoryContentGetOptions{},
			)

			defer downloadedFile.Close()
			fileContent, err := io.ReadAll(downloadedFile)
			if err != nil {
				w.WriteHeader(500)
				fmt.Printf("error: %s\n", err)
				return
			}

			_, _, err = deploy(ctx, s.client, fileContent)
			if err != nil {
				w.WriteHeader(500)
				fmt.Printf("error: %s\n", err)
				return
			}
			fmt.Printf("Deployed %s\n", file)
		}
	default:
		w.WriteHeader(500)
		fmt.Printf("unknown event type %s\n", event)
	}
}

func getFiles(commits []*github.HeadCommit) []string {

	allFiles := []string{}
	for _, commit := range commits {
		allFiles = append(allFiles, commit.Added...)
		allFiles = append(allFiles, commit.Modified...)
	}

	allUniqueFiles := make(map[string]bool)
	for _, file := range allFiles {
		allUniqueFiles[file] = true
	}
	allUniqueFilesSlice := []string{}
	for file := range allUniqueFiles {
		allUniqueFilesSlice = append(allUniqueFilesSlice, file)
	}
	return allUniqueFilesSlice
}
