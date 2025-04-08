package translator

type IClient interface {
	Translate(texts []string) ([]string, error)
}
