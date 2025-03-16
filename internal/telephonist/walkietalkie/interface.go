package walkietalkie

type IController interface {
	Talk(userID string, key, input string) (asnwer string)
	Stop(userID, key string) (id string, err error)
}
