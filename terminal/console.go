package terminal

type Console struct {
	*console
}
type console struct {
}

func InitializeConsole() *Console {
	return &Console{}
}

func (self *console) Render() {

}
