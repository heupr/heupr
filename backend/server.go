package backend

// Server hosts backend in-memory active repos and access to the database.
type Server struct {
	database dataAccess
	repos    map[int64]*repo
}

// Start is exported so that cmd/ has access to launch the backend.
func (s *Server) Start() {
	// read settings + installation/integration info from database
	// call parseSettings and pass in settings
	// call newRepo for each repo
	// call timer
}

func (s *Server) timer() {
	// start ticker
	// spin up workers/dispatchers
	// start perpetual goroutine
	// select case
	// ticker.C
	// - call s.database.read() to return container map
	// - pass resulting container map into the collector
	// - ^ possibly truncate the collector function into in-timer logic
	// ender
	// - close ender
	// stop ticker
}
