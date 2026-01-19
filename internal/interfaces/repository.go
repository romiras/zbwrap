package interfaces

// RepositoryManager defines the contract for managing ZBackup repositories
type RepositoryManager interface {
	Add(alias, path string) error
	Get(alias string) (string, bool)
	List() map[string]string
}
