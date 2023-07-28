package binder

type Portal struct {
	document Document
	buffer   OTBuffer
	// config   Config

	exitChan chan<- 
}

// type Config struct {
// 	otConfig OTBufferConfig
// }
