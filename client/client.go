package client

type Client interface {
	Hosts() []string

	GetFile(host string, filename string) error
	ListFiles(host string) ([]RemoteFile, error)

	Shared() []LocalFile
	ShareFile(name, filepath string)
}

type RemoteFile struct {
	Name string `json:"name"`
	Mime string `json:"mime"`
}

type LocalFile struct {
	Name, Path, Mime string
}