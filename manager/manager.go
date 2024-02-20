package manager

type manager struct {
	storageDir string
}

func NewManager(certStore string) *manager {
	return &manager{
		storageDir: certStore,
	}
}
