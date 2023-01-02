package terminalReader

func New() TerminalReader {
	t := terminalReader{}
	t.parse()
	return &t
}
