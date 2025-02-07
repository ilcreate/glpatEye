package gitlabClient

import "context"

type GitlabClient struct {
	Token    string
	BaseURL  string
	Pattern  string
	Response Response
}

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	Projects Projects `json:"projects"`
	Groups   Projects `json:"groups"`
}

type Projects struct {
	Nodes    []ProjectNode `json:"nodes"`
	PageInfo PageInfo      `json:"pageInfo"`
}

type ProjectNode struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	HttpUrlToRepo string `json:"httpUrlToRepo"`
	WebUrl        string `json:"webUrl"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type AccessToken struct {
	Name        string `json:"name"`
	ID          int    `json:"id"`
	LastUsed    string `json:"last_used_at"`
	ExpiresAt   string `json:"expires_at"`
	ProjectName string `json:"project_name"`
	UrlToRepo   string `json:"url_to_repo"`
	DaysExpire  int    `json:"days_until_expire"`
}

type CheckSignature struct {
	Ctx           context.Context
	Client        *GitlabClient
	ResourceType  string
	IDs           chan string
	ObjectMap     map[string]ProjectNode
	ResultChannel chan []AccessToken
	ErrorChannel  chan error
	Counter       *int32
}
