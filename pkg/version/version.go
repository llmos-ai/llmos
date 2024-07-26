package version

var (
	Version = "v0.0.0-dev"
	Commit  = "HEAD"
)

func GetFriendlyVersion() string {
	if Commit == "" {
		return Version
	}
	return Version + "-" + Commit
}
