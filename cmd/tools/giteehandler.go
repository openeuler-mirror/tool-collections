/*
Copyright 2019 The openeuler community Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gitee.com/openeuler/go-gitee/gitee"
	"golang.org/x/oauth2"
)

var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

type SimplifiedRepo struct {
	Id        int     `json:"id,omitempty"`
	FullName  *string `json:"full_name,omitempty"`
	Url       *string  `json:"url,omitempty"`
}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

type GiteeHandler struct {
	GiteeClient *gitee.APIClient
	Token string
	Context context.Context
}


func NewGiteeHandler(giteeToken string) *GiteeHandler{
	// oauth
	oauthSecret := checkOwnerFlags.GiteeToken
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)

	// configuration
	giteeConf := gitee.NewConfiguration()
	giteeConf.HTTPClient = oauth2.NewClient(ctx, ts)

	// git client
	giteeClient := gitee.NewAPIClient(giteeConf)
	return &GiteeHandler{
		GiteeClient:giteeClient,
		Token: giteeToken,
		Context: ctx,
	}
}

func (gh *GiteeHandler) ValidateUser(wg *sync.WaitGroup, stopChannel <-chan struct{}, users <-chan string, invalid *[]string) {
	defer wg.Done()
	for {
		select {
		case u, ok := <- users:
			if !ok {
				fmt.Printf("User channel finished, quiting..\n")
				return
			} else {
				if !gh.checkUserExists(u) {
					*invalid = append(*invalid, u)
				}
			}
		case <-stopChannel:
			fmt.Println("quit signal captured, quiting.")
			return
		}
	}
}

func (gh *GiteeHandler) ValidateRepo(wg *sync.WaitGroup, stopChannel <-chan struct{}, rsChannel chan<- string, repos <-chan string, orgName string) {
	defer wg.Done()
	for {
		select {
		case u, ok := <- repos:
			if !ok {
				fmt.Printf("Repo channel finished, quiting..\n ")
				return
			} else {
				if !gh.checkRepoExists(u, orgName) {
					rsChannel <- u
				}
			}
		case <-stopChannel:
			fmt.Println("quit signal captured, quiting.")
			return
		}
	}
}

func (gh *GiteeHandler) checkUserExists(name string) bool {
	option := gitee.GetV5UsersUsernameOpts{
		AccessToken: optional.NewString(gh.Token),
	}
	_, result, err := gh.GiteeClient.UsersApi.GetV5UsersUsername(gh.Context, name, &option)
	if err != nil {
		if result.StatusCode == 404 {
			fmt.Printf("[Warning] User %s does not exists. \n", name)
			return false
		} else {
			fmt.Printf("Failed to recognize user %s on gitee website, skipping", name)
		}
	}
	return true
}

func (gh *GiteeHandler) checkRepoExists(name, orgName string) bool {
	option := gitee.GetV5SearchRepositoriesOpts{
		AccessToken: optional.NewString(gh.Token),
		Owner: optional.NewString(orgName),
	}
	projects, result, err := gh.GiteeClient.SearchApi.GetV5SearchRepositories(gh.Context, name, &option)
	if err != nil || result.StatusCode != 200 {
		fmt.Printf("[Warning] Repo %s does not exists, or failed %v \n", name, err)
		return false
	}
	//check total count
	if len(projects) == 0 {
		fmt.Printf("[Warning] can't found project %s from gitee website\n", name)
		return false
	}

	var projectNames []string
	for _, p := range projects {
		if p.FullName == name {
			return true
		}
		projectNames = append(projectNames, p.FullName)
	}
	fmt.Printf("[Warning] Unable to get repo %s information, actual search result %s \n", name, strings.Join(projectNames, ","))
	return false
}

