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
	Alias   string `json:"alias"`
	BaseURL string `json:"base_url"`
}
