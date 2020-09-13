package serializer

const (
	VCSTypeGithub    = "github"
	VCSTypeBitbucket = "bitbucket"
)

type VCS struct {
	Alias   string `json:"alias"`
	BaseURL string `json:"base_url"`
}
