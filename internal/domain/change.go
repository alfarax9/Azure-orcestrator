package domain

// Change is one fact that changed: exactly one field is set, holding the object as it is
// now. The store reports a Change after every successful write, and the API passes them to
// clients, which replace the object with the same id.
type Change struct {
	Workspace *Workspace
	Project   *Project
	Worker    *Worker
}

// Device is a client that paired with the daemon, such as a desktop app or a browser. The
// token that lets it in is never kept, only its hash.
type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt Timestamp `json:"createdAt"`
	LastSeen  Timestamp `json:"lastSeen"`
}
