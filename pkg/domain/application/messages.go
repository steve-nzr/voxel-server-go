package application

// An Actor message containing some data received or sent from and to the client socket.
type DataExchangeMessage struct {
	Data []byte
}
