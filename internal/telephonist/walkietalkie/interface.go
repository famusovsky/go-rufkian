package walkietalkie

//go:generate mockgen -package walkietalkie -mock_names IController=ControllerMock -source ./interface.go -typed -destination interface.mock.gen.go
type IController interface {
	Talk(userID, key, input string) (asnwer string)
	Stop(userID string) (id string, err error)
	CleanUp()
}
