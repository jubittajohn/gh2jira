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

package config

import (
	"errors"
	"os"
	"fmt"

	"sigs.k8s.io/yaml"
)

const schemaName string = "gh2jira.config"
const defaultJiraBaseURL string = "https://gh2jiratest.atlassian.net/"

type tokenStore struct {
	JiraToken   string `json:"jira"`
	GithubToken string `json:"github"`
}

type Config struct {
	Schema      string     `json:"schema"`
	JiraBaseUrl string     `json:"jiraBaseUrl,omitempty"`
	JiraUsername	string
	Tokens      tokenStore `json:"authTokens"`
}

func (c *Config) setDefaults() error {
	if c.JiraBaseUrl == "" {
		c.JiraBaseUrl = defaultJiraBaseURL
	}
	return nil
}

func ReadFile(f string) (*Config, error) {
	b, err := readFile(f)
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	if c.Schema != schemaName {
		return nil, fmt.Errorf("invalid schema: %q should be %q: %v", c.Schema, schemaName, err)
	}
	if c.Tokens.GithubToken == "" {
		return nil, errors.New("missing required github token")
	}
	if c.Tokens.JiraToken == "" {
		return nil, errors.New("missing required jira token")
	}
	if c.JiraUsername == "" {
		return nil, errors.New("missing required jira username")
	}

	c.setDefaults()

	return &c, nil
}

// overrideable func for mocking os.ReadFile
var readFile = func(file string) ([]byte, error) {
	return os.ReadFile(file)
}
