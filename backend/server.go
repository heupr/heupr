package backend

// Server hosts backend in-memory active repos and access to the database.
type Server struct {
	database dataAccess
	repos    map[int64]*repo
}

// Start is exported so that cmd/ has access to launch the backend.
func (s *Server) Start() {
	// open database and initialize the s.repos field
	// read integrations, settings, and events into memory
	// loop over map + instantiate repo structs into s.repos field
	// reference settings + s.repos value using integrations map key
	// establish response paths + initialize logic per repo
	// reference events slice using integrations map key
	// pass into necessary learn methods per response
	// start timer() method
	// note a sync.WaitGroup may be useful if these actions are piped into
	// channels/goroutines
}

func (s *Server) timer(ender chan bool) {
	// start ticker + dispatcher
	// begin perpetual goroutine
	// if ticker.C
	// read new integrations, settings, and events into memory
	// - place into work struct containing all three
	// - pass work struct into collector
	// if ender
	// - stop ticker, close ender, and return
}
