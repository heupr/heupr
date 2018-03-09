package backend

// Server hosts backend in-memory active repos and access to the database.
type Server struct {
	database accessDB
	repos    map[int64]*repo
}

// Start is exported so that cmd/ has access to launch the backend.
func (s *Server) Start() {
	// read settings + installation/integration info from database
	// loop over settings
	// set settings into each repo settings field
	// call newRepo for each repo
	// call timer
}

func (s *Server) timer() {
	// NOTE: see the existing code; it will be nearly identical in all likelihood
}
