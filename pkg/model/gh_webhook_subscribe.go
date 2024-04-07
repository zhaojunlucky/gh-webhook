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
	Event        string   // mandatory
	Actions      []string // optional
	OrgRepo      string   // optional, null to match all, regex
	PullRequest  PullRequestSubscribe
	Push         PushSubscribe
	IssueComment IssueComment
	Expr         string // https://expr-lang.org/docs/configuration
}
