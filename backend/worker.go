package backend

import (
	"heupr/backend/response/preprocess"
)

type worker struct {
	id    int
	work  chan *preprocess.Container
	queue chan chan *preprocess.Container
	repos map[int64]*repo
	quit  chan bool
}

func (w *worker) start() {
	// start perpetual goroutine
	// select case
	// receive work from work chan
	// - if push event
	// - - if payload contains .heupr.toml file
	// - - - send request to GitHub using GetContents method
	// - - - pull out Content field from results and decode base64
	// - - - pass string(results) into the BurntSushi/toml library to get TOML struct
	// - - - populate into respective repo settings field
	// - - - call parseSettings method
	// - else if other event
	// - - process to retrieve event-action combo
	// - - use to populate into respective map value models/conditionals
	// quit
	// - return
}
