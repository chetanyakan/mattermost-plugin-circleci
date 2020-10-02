package serializer

const (
	VCSTypeGithub    = "github"
	VCSTypeBitbucket = "bitbucket"
)

var (
	DefaultVCSList = map[string]*VCS{
		"github": {
			Alias:   "github",
			Type:    VCSTypeGithub,
			BaseURL: "https://github.com",
		},
		"bitbucket": {
			Alias:   "bitbucket",
			Type:    VCSTypeBitbucket,
			BaseURL: "https://bitbucket.org",
		},
	}
)

type VCS struct {
	Alias   string `json:"alias"`
	Type    string `json:"type"`
	BaseURL string `json:"base_url"`
}
