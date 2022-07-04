package src

import "syscall/js"

func AddWasm(this js.Value, args []js.Value) any {
	return Add(args[0].Int(), args[1].Int())
}
func GreetWasm(this js.Value, args []js.Value) any {
	Greet(args[0].String())
	return nil
}
func mainWasm() {
	js.Global().Set("Add", js.FuncOf(AddWasm))
	js.Global().Set("Greet", js.FuncOf(GreetWasm))
}
