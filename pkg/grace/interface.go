package grace

type Process interface {
	Run()
	Shutdown()
}
