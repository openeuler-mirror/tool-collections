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
	"strconv"
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
				fmt.Printf("Starting to validate user %s\n", u)
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

func (gh *GiteeHandler) checkUserExists(name string) bool {
	option := gitee.GetV5UsersUsernameOpts{
		AccessToken: optional.NewString(gh.Token),
	}
	_, result, err := gh.GiteeClient.UsersApi.GetV5UsersUsername(gh.Context, name, &option)
	if err != nil {
		if result != nil && result.StatusCode == 404 {
			fmt.Printf("[Warning] User %s does not exists. \n", name)
			return false
		} else {
			fmt.Printf("Failed to recognize user %s on gitee website, skipping", name)
		}
	}
	return true
}

func (gh *GiteeHandler) CollectRepoPageCount(pageSize int, enterpriseName string) int {
	option := gitee.GetV5EnterprisesEnterpriseReposOpts{
		AccessToken: optional.NewString(gh.Token),
		PerPage: optional.NewInt32(int32(pageSize)),
	}
	_, result, err := gh.GiteeClient.RepositoriesApi.GetV5EnterprisesEnterpriseRepos(gh.Context, enterpriseName, &option)
	if err != nil || result.StatusCode != 200 {

		fmt.Printf("[Error] Can't collect projects in enterprise %s, %v \n", enterpriseName, err)
		return -1
	}
	size, ok := result.Header["Total_page"]
	if !ok {
		fmt.Printf("[Error] Can't collect 'Total_page' from Header %v", result.Header)
		return -1
	}
	sizeInt, err := strconv.ParseInt(size[0], 10, 0)
	if err != nil {
		fmt.Printf("[Error] Can't convert 'Total_page' to integer %v", size)
		return -1
	}
	return int(sizeInt)
}

func (gh *GiteeHandler) CollectRepos(wg *sync.WaitGroup, pageSize, totalPage, workerIndex, gap int, enterpriseName string, rsChannel chan<- string) {
	defer wg.Done()
	for i := workerIndex; i <= totalPage; i+=gap {
		fmt.Printf("Starting to fetch project lists %d/%d from enterpise %s\n", i, totalPage, enterpriseName)
		option := gitee.GetV5EnterprisesEnterpriseReposOpts{
			AccessToken: optional.NewString(gh.Token),
			PerPage: optional.NewInt32(int32(pageSize)),
			Page: optional.NewInt32(int32(i)),
		}
		projects, result, err := gh.GiteeClient.RepositoriesApi.GetV5EnterprisesEnterpriseRepos(gh.Context, enterpriseName, &option)
		if err != nil || result.StatusCode != 200 {
			fmt.Printf("[Warning] Failed to get projects %d/%d from enterprise %s\n", i, totalPage, enterpriseName)
			continue
		}
		for _,p := range projects {
			rsChannel <- p.FullName
		}
	}
}

