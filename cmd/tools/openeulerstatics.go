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
}



var openEulerStatics = &OpenEulerStatics{}

func InitStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&openEulerStatics.GiteeToken, "giteetoken", "g", "", "the gitee token")
	cmd.Flags().StringVarP(&openEulerStatics.HWUser, "huaweicloud user", "u", "", "the username for huawei cloud")
	cmd.Flags().StringVarP(&openEulerStatics.HWPassword, "huaweiclud password", "p", "", "the password for huawei cloud")
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
	err := ShowStatics()
	if err != nil {
		return err
	}
	err = ShowStatics2()
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

func ShowStatics2() error {
	var collectingwg sync.WaitGroup
	var endwg sync.WaitGroup
	var totalUsers []string
	var totalSubscribeUsers []string
	var totalProjects []string
	resultChannel := make(chan string, 50)
	subscribeChannel := make(chan string, 50)
	projectChannel := make(chan string, 50)
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
	for i := 1; i <= 5; i++ {
		collectingwg.Add(1)
		go giteeHandler.CollectRepos(&collectingwg,100, size, i, 5 , "open_euler", projectChannel, )
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

	for i := 0; i <= len(totalProjects); i+=10 {
		groupwg := sync.WaitGroup{}
		for j := i; j <= i+9; j++ {

			if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], "src-openeuler/")) {
				fmt.Printf("Collecting Star info for project %s\n", totalProjects[j])
				groupwg.Add(1)
				go giteeHandler.ShowRepoStarStatics(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], resultChannel)
			}
		}
		groupwg.Wait()
	}

	for i := 0; i <= len(totalProjects); i+=10 {
		groupwg := sync.WaitGroup{}
		for j := i; j <= i+9; j++ {

			if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], "src-openeuler/")) {
				fmt.Printf("Collecting Subsribe info for project %s\n", totalProjects[j])
				groupwg.Add(1)
				go giteeHandler.ShowRepoWatchStatics(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], subscribeChannel)
			}
		}
		groupwg.Wait()
	}

	close(resultChannel)
	close(subscribeChannel)
	endwg.Wait()
	fmt.Printf("[Result] There are %d users stars src-openeuler project \n: %s \n", len(totalUsers), "")
	fmt.Printf("[Result] There are %d users subscribe src-openeuler project \n: %s \n", len(totalSubscribeUsers), "")
	return nil
}


func ShowStatics() error {
	var collectingwg sync.WaitGroup
	var endwg sync.WaitGroup
	var totalUsers []string
	var totalSubscribeUsers []string
	var totalProjects []string
	resultChannel := make(chan string, 50)
	subscribeChannel := make(chan string, 50)
	projectChannel := make(chan string, 50)
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
	for i := 1; i <= 5; i++ {
		collectingwg.Add(1)
		go giteeHandler.CollectRepos(&collectingwg,100, size, i, 5 , "open_euler", projectChannel, )
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

	for i := 0; i <= len(totalProjects); i+=10 {
		groupwg := sync.WaitGroup{}
		for j := i; j <= i+9; j++ {

			if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], "openeuler/")) {
				fmt.Printf("Collecting Star info for project %s\n", totalProjects[j])
				groupwg.Add(1)
				go giteeHandler.ShowRepoStarStatics(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], resultChannel)
			}
		}
		groupwg.Wait()
	}

	for i := 0; i <= len(totalProjects); i+=10 {
		groupwg := sync.WaitGroup{}
		for j := i; j <= i+9; j++ {

			if  j < len(totalProjects) && (strings.HasPrefix(totalProjects[j], "openeuler/")) {
				fmt.Printf("Collecting Subsribe info for project %s\n", totalProjects[j])
				groupwg.Add(1)
				go giteeHandler.ShowRepoWatchStatics(&groupwg, strings.Split(totalProjects[j], "/")[0], strings.Split(totalProjects[j], "/")[1], subscribeChannel)
			}
		}
		groupwg.Wait()
	}

	close(resultChannel)
	close(subscribeChannel)
	endwg.Wait()
	fmt.Printf("[Result] There are %d users stars openeuler project \n: %s \n", len(totalUsers), "")
	fmt.Printf("[Result] There are %d users subscribe openeuler project \n: %s \n", len(totalSubscribeUsers), "")
	return nil
}



