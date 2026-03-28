package application

type Options struct {
	Server Options_Server
	Mojang Options_Mojang
}

type Options_Server struct {
	Address        string
	Port           int
	ConnBufferSize int
}

type Options_Mojang struct {
	SessionServerURL string
}
