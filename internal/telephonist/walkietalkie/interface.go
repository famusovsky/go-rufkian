package walkietalkie

type IController interface {
	Talk(userID string, key, input string) (asnwer string)
	Stop(userID string) (id string, err error)
}
