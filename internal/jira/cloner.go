// Copyright © 2022 jesus m. rodriguez jmrodri@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jira

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gojira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/google/go-github/v47/github"
)

type Option func(*ClonerConfig) error

type ClonerConfig struct {
	client       *http.Client
	token        string
	dryRun       bool
	project      string
	jiraBaseURL  string
	jiraUsername string
}

func (c *ClonerConfig) setDefaults() error {
	if c.client == nil {
		if c.token == "" {
			return errors.New("cannot create jira client without a token")
		}
		fmt.Printf("username %v \n",c.jiraUsername)
		tp := &gojira.BasicAuthTransport{
			APIToken: c.token,
			Username: c.jiraUsername,
		}
		c.client = tp.Client()
	}
	return nil
}

func WithClient(cl *http.Client) Option {
	return func(c *ClonerConfig) error {
		c.client = cl
		return nil
	}
}

func WithToken(token string) Option {
	return func(c *ClonerConfig) error {
		c.token = token
		return nil
	}
}

func WithDryRun(dr bool) Option {
	return func(c *ClonerConfig) error {
		c.dryRun = dr
		return nil
	}
}

func WithProject(p string) Option {
	return func(c *ClonerConfig) error {
		c.project = p
		return nil
	}
}

func WithJiraBaseURL(j string) Option {
	return func(c *ClonerConfig) error {
		c.jiraBaseURL = j
		return nil
	}
}

func WithJiraUsername(u string) Option {
	return func(c *ClonerConfig) error {
		c.jiraUsername = u
		return nil
	}
}

func getWebURL(url string) string {
	if url == "" {
		return url
	}
	return strings.Replace(strings.Replace(url, "api.github.com", "github.com", 1), "repos/", "", 1)
}

func Clone(issue *github.Issue, opts ...Option) (*gojira.Issue, error) {
	config := ClonerConfig{}
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	if err := config.setDefaults(); err != nil {
		return nil, err
	}

	// actually send issue create command
	jiraClient, err := gojira.NewClient(config.jiraBaseURL, config.client)
	if err != nil {
		return nil, err
	}

	ji := gojira.Issue{
		Fields: &gojira.IssueFields{
			Description: fmt.Sprintf("%s\n\nUpstream Github issue: %s\n", issue.GetBody(), getWebURL(issue.GetURL())),
			Type: gojira.IssueType{
				Name: "Story",
			},
			Project: gojira.Project{
				Key: config.project,
			},
			Summary: fmt.Sprintf("[UPSTREAM] %s #%d", issue.GetTitle(), issue.GetNumber()),
		},
	}

	var createdIssue *gojira.Issue

	if config.dryRun {
		fmt.Println("\n############# DRY RUN MODE #############")
		fmt.Printf("Cloning issue #%d to jira project board: %s\n\n", issue.GetNumber(), ji.Fields.Project.Key)
		fmt.Printf("Summary: %s\n", ji.Fields.Summary)
		fmt.Printf("Type: %s\n", ji.Fields.Type.Name)
		fmt.Println("Description:")
		fmt.Printf("%s\n", ji.Fields.Description)
		fmt.Println("\n############# DRY RUN MODE #############")
	} else {
		fmt.Printf("Creating new issue \n")
		fmt.Printf("Cloning issue #%d to jira project board: %s\n\n", issue.GetNumber(), ji.Fields.Project.Key)
		var err error
		// actually send issue create command
		createdIssue, _, err = jiraClient.Issue.Create(context.Background(), &ji)
		if err != nil {
			fmt.Printf("Error cloning issue: %v", err.Error())
			return createdIssue, err
		}

		if createdIssue != nil {
			fmt.Printf("Issue cloned; see %s\n",
				fmt.Sprintf(config.jiraBaseURL+"browse/%s", createdIssue.Key))
		}
	}

	return createdIssue, nil
}
