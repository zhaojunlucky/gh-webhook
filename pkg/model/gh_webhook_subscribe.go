package model

type PullRequestSubscribe struct {
	AllowedBaseBranches    []string // optional, null to match all, regex
	DisallowedBaseBranches []string // optional, null to match nothing, regex, high priority than AllowedBranches
}

type PushSubscribe struct {
	AllowedPushRefs []string // optional, null to match all, regex
}

type IssueComment struct {
	AllowedComments []string // optional, null to match all, regex
}

type GHWebHookSubscribe struct {
	Event        string               // mandatory
	Actions      []string             `gorm:"serializer:json"` // optional
	OrgRepo      string               // optional, null to match all, regex
	PullRequest  PullRequestSubscribe `gorm:"serializer:json"`
	Push         PushSubscribe        `gorm:"serializer:json"`
	IssueComment IssueComment         `gorm:"serializer:json"`
	Expr         string               // https://expr-lang.org/docs/configuration
}
