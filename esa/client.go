package esa

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	Endpoint   = "api.esa.io"
	Stylesheet = "https://assets.esa.io/assets/application-860deb72f57963abb3cecce7b8070ab4e106b68cee8e3205d457110507b494f4.css"
)

type Author struct {
	Myself     bool
	Name       string
	ScreenName string `json:"screen_name"`
	Icon       string
}

type Post struct {
	Number         int
	Name           string
	FullName       string `json:"full_name"`
	Wip            bool
	BodyMd         string        `json:"body_md"`
	BodyHtml       template.HTML `json:"body_html"`
	CreatedAt      string        `json:"created_at"`
	Message        string
	Url            string
	UpdatedAt      string `json:"updated_at"`
	Tags           []string
	Category       string
	RevisionNumber int    `json:"revision_number"`
	CreatedBy      Author `json:"created_by"`
	UpdatedBy      Author `json:"updated_by"`
}

type Posts struct {
	Posts      []Post
	PrevPage   int `json:"prev_page"`
	NextPage   int `json:"next_page"`
	TotalCount int `json:"total_count"`
	Page       int
	PerPage    int `json:"per_page"`
	MaxPerPage int `json:"max_per_page"`
}

type Client struct {
	Team    string
	PerPage int
	Http    *http.Client
}

func NewClient(team string, perPage int) *Client {
	httpCli := &http.Client{}

	cli := &Client{
		Team:    team,
		PerPage: perPage,
		Http:    httpCli,
	}

	return cli
}

func (cli *Client) Posts(token string, q string, page string) (*Posts, error) {
	req, err := buildRequest(token, cli.Team, q, page, strconv.Itoa(cli.PerPage))

	if err != nil {
		return nil, err
	}

	res, err := cli.Http.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", res.Status, body)
	}

	posts := &Posts{}
	err = json.Unmarshal(body, &posts)

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func buildRequest(token, team, q, page, perPage string) (*http.Request, error) {
	url := fmt.Sprintf("https://%s/v1/teams/%s/posts", Endpoint, team)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("sort", "number")
	query.Add("order", "desc")
	query.Add("per_page", perPage)

	if q != "" {
		query.Add("q", q)
	}

	if page != "" {
		query.Add("page", page)
	}

	req.URL.RawQuery = query.Encode()
	req.Header.Add("Authorization", "Bearer "+token)

	return req, nil
}
