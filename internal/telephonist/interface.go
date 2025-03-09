package telephonist

type IServer interface {
	Run()
	Shutdown()
}
