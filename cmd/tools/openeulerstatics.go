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
	"fmt"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/spf13/cobra"
	"strings"
	"sync"
	"time"
)

type OpenEulerStatics struct {
	GiteeToken string
	HWUser string
	HWPassword string
	ShowStar bool
	ShowSubscribe bool
	ShowPR bool
	Threads int32
}



var openEulerStatics = &OpenEulerStatics{}

func InitStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&openEulerStatics.GiteeToken, "giteetoken", "g", "", "the gitee token")
	cmd.Flags().StringVarP(&openEulerStatics.HWUser, "user", "u", "", "the username for huawei cloud")
	cmd.Flags().StringVarP(&openEulerStatics.HWPassword, "password", "p", "", "the password for huawei cloud")
	cmd.Flags().BoolVarP(&openEulerStatics.ShowStar, "showstar", "s", false, "whether show stars count")
	cmd.Flags().BoolVarP(&openEulerStatics.ShowSubscribe, "showsubscribe", "w", false, "whether show subscribe count")
	cmd.Flags().BoolVarP(&openEulerStatics.ShowPR, "showpr", "r", false, "whether show pr count")
	cmd.Flags().Int32VarP(&openEulerStatics.Threads, "threads", "t", 5, "how many threads to perform")
}

func buildStaticsCommand() *cobra.Command {
	staticsCommand := &cobra.Command{
		Use:   "statics",
		Short: "show current statics of openEuler",
	}

	showCommand := &cobra.Command{
		Use:   "show",
		Short: "show current statics of openEuler",
		Run: func(cmd *cobra.Command, args []string) {
			checkError(cmd, ShowAllStatics())
		},
	}
	InitStaticsFlags(showCommand)
	staticsCommand.AddCommand(showCommand)

	return staticsCommand
}

func ShowAllStatics() error {
	fmt.Printf("now is %s\n", time.Now().String())
	err := ShowStatics("openeuler")
	if err != nil {
		return err
	}
	ShowStatics("src-openeuler")
	return nil
}

func ShowObsStatics() error {
	authUrl := "https://iam.cn-south-1.myhuaweicloud.com/v3"
	region := "cn-south-1"
	client := NewHWCloudHandler(authUrl, openEulerStatics.HWUser, openEulerStatics.HWPassword, region)
	_, err := openstack.NewCESV1(client.Client, golangsdk.EndpointOpts{
	})
	if err != nil {
		return err
	}

	return nil
}

func ShowStatics(organization string) error {
	var collectingwg sync.WaitGroup
	var endwg sync.WaitGroup
	var totalUsers []string
	var totalSubscribeUsers []string
	var totalProjects []string
	resultChannel := make(chan string, 50)
	subscribeChannel := make(chan string, 50)
	projectChannel := make(chan string, 50)
	prChannel := make(chan PullRequest, 50)
	prResults := []PullRequest{}
	// Collecting contributing information from openeuler organization
	giteeHandler := NewGiteeHandler(openEulerStatics.GiteeToken)
	// Running 5 workers to collect the projects status
	size := giteeHandler.CollectRepoPageCount(100, "open_euler")
	if size <= 0 {
		return fmt.Errorf("can't get any projects in enterprise 'open_euler'")
	}

	go func() {
		endwg.Add(1)
		for rs := range projectChannel {
			totalProjects = append(totalProjects, rs)
		}
		endwg.Done()
	}()
	for i := 1; i <= int(openEulerStatics.Threads); i++ {
		collectingwg.Add(1)
		go giteeHandler.CollectRepos(&collectingwg,100, size, i, int(openEulerStatics.Threads) , "open_euler", projectChannel, )
	}

	collectingwg.Wait()
	close(projectChannel)
	endwg.Wait()

	go func() {
		endwg.Add(1)
		for rs := range resultChannel {
			if !Find(totalUsers, rs) {
				totalUsers = append(totalUsers, rs)
			}
		}
		endwg.Done()
	}()

	go func() {
		endwg.Add(1)
		for rs := range subscribeChannel {
			if !Find(totalSubscribeUsers, rs) {
				totalSubscribeUsers = append(totalSubscribeUsers, rs)
			}
		}
		endwg.Done()
	}()

	go func() {
		endwg.Add(1)
		for pr := range prChannel {
			prResults = append(prResults, pr)
		}
		endwg.Done()
	}()

	if (openEulerStatics.ShowStar) {
		for i := 0; i <= len(totalProjects); i+=int(openEulerStatics.Threads){
			groupwg := sync.WaitGroup{}
			for j := i; j < i+int(openEulerStatics.Threads); j++ {

				if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], fmt.Sprintf("%s/", organization))) {
					fmt.Printf("Collecting Star info for project %s\n", totalProjects[j])
					groupwg.Add(1)
					go giteeHandler.ShowRepoStarStatics(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], resultChannel)
				}
			}
			groupwg.Wait()
		}
	}

	if (openEulerStatics.ShowSubscribe){
		for i := 0; i <= len(totalProjects); i+=int(openEulerStatics.Threads) {
			groupwg := sync.WaitGroup{}
			for j := i; j < i+int(openEulerStatics.Threads); j++ {

				if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], fmt.Sprintf("%s/", organization))) {
					fmt.Printf("Collecting Subsribe info for project %s\n", totalProjects[j])
					groupwg.Add(1)
					go giteeHandler.ShowRepoWatchStatics(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], subscribeChannel)
				}
			}
			groupwg.Wait()
		}
	}

	if (openEulerStatics.ShowPR) {
		for i := 0; i <= len(totalProjects); i+=int(openEulerStatics.Threads) {
			groupwg := sync.WaitGroup{}
			for j := i; j <= i+int(openEulerStatics.Threads); j++ {

				if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], fmt.Sprintf("%s/", organization))) {
					fmt.Printf("Collecting PR info for project %s\n", totalProjects[j])
					groupwg.Add(1)
					go giteeHandler.ShowRepoPRs(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], prChannel)
				}
			}
			groupwg.Wait()
		}
	}

	close(resultChannel)
	close(subscribeChannel)
	close(prChannel)
	endwg.Wait()
	if (openEulerStatics.ShowStar) {
		fmt.Printf("[Result] There are %d users stars %s project \n.", len(totalUsers), organization)
	}
	if (openEulerStatics.ShowSubscribe) {
		fmt.Printf("[Result] There are %d users subscribe %s project \n.", len(totalSubscribeUsers), organization)
	}
	if (openEulerStatics.ShowPR) {
		fmt.Printf("[Result] The contribution info for  %sis: \n", organization)
		fmt.Printf("Repo, CreateAt, PR Number, Auther, State, Link\n")
	}
	for _,pr := range prResults {
		fmt.Printf("%s, %s, %d, %s, %s, %s\n", pr.RepoName, pr.CreateAt, pr.Number, pr.Auther, pr.State, pr.Link)
	}
	return nil
}
