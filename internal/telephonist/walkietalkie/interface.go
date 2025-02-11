package walkietalkie

type IController interface {
	Talk(key string, input string) (asnwer string)
	Stop(key string) (id uint64, err error)
}
