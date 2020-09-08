package serializer

const (
	VCSTypeGithub    = "github"
	VCSTypeBitbucket = "bitbucket"
)

var (
	validVCSTypes = map[string]bool{
		VCSTypeGithub:    true,
		VCSTypeBitbucket: true,
	}
)

type VCS struct {
	Type    string `json:"type"`
	BaseURL string `json:"base_url"`
}
