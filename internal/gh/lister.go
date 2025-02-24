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

package gh

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

type Option func(*ListerConfig) error

type ListerConfig struct {
	client    *http.Client
	Milestone string
	Token     string
	Assignee  string
	Project   string
	Label     []string
}

func (c *ListerConfig) setDefaults() error {
	if c.client == nil {
		ctx := context.Background()
		if c.Token == "" {
			return errors.New("cannot create github client without a token")
		}
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: c.Token},
		)
		c.client = oauth2.NewClient(ctx, ts)
	}
	return nil
}

func (l *ListerConfig) GetGithubOrg() string {
	return strings.Split(l.Project, "/")[0]
}

func (l *ListerConfig) GetGithubRepo() string {
	s := strings.Split(l.Project, "/")
	if len(s) == 1 {
		return s[0]
	}
	return s[1]
}

func WithClient(cl *http.Client) Option {
	return func(c *ListerConfig) error {
		c.client = cl
		return nil
	}
}

func WithToken(token string) Option {
	return func(c *ListerConfig) error {
		c.Token = token
		return nil
	}
}

func WithMilestone(m string) Option {
	return func(c *ListerConfig) error {
		c.Milestone = m
		return nil
	}
}

func WithAssignee(a string) Option {
	return func(c *ListerConfig) error {
		c.Assignee = a
		return nil
	}
}

func WithProject(p string) Option {
	return func(c *ListerConfig) error {
		c.Project = p
		return nil
	}
}

func WithLabel(l []string) Option {
	return func(c *ListerConfig) error {
		c.Label = l
		return nil
	}
}

func GetIssue(issueNum int, opts ...Option) (*github.Issue, error) {
	config := ListerConfig{}
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	if err := config.setDefaults(); err != nil {
		return nil, err
	}

	client := github.NewClient(config.client)

	issue, _, err := client.Issues.Get(context.Background(), config.GetGithubOrg(),
		config.GetGithubRepo(), issueNum)

	if err != nil {
		return nil, err
	}
	return issue, nil
}

func ListIssues(opts ...Option) ([]*github.Issue, error) {
	config := ListerConfig{}
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	if err := config.setDefaults(); err != nil {
		return nil, err
	}

	client := github.NewClient(config.client)

	opt := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 50},
		State:       "open",
		Milestone:   config.Milestone,
		Assignee:    config.Assignee,
		Labels:      config.Label,
	}

	var allIssues []*github.Issue

	for {
		issues, resp, err := client.Issues.ListByRepo(context.Background(),
			config.GetGithubOrg(), config.GetGithubRepo(), opt)

		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allIssues, nil
}
