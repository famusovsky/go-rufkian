package translator

//go:generate mockgen -package translator -mock_names IClient=ClientMock -source ./interface.go -typed -destination interface.mock.gen.go
type IClient interface {
	Translate(texts []string) ([]string, error)
}
