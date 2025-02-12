package walkietalkie

type IController interface {
	Talk(userID uint64, key, input string) (asnwer string)
	Stop(userID uint64) (id uint64, err error)
}
