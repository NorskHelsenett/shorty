package models

// Redirect represents a URL redirection with a short path
type Redirect struct {
	Path string `json:"path,omitempty"` // key/id
	URL  string `json:"url,omitempty"`
}

// RedirectUser represents a user with permission to create redirects
type RedirectUser struct {
	Email string `json:"email"`
}

// RedirectPath represents a redirect with ownership information
type RedirectPath struct {
	Path  string `json:"path,omitempty"`
	URL   string `json:"url,omitempty"`
	Owner string `json:"owner,omitempty"`
}

// RedirectAllPaths represents a redirect with ownership and permissions
type RedirectAllPaths struct {
	Path   string `json:"path,omitempty"`
	URL    string `json:"url,omitempty"`
	Owner  string `json:"owner,omitempty"`
	Modify bool   `json:"modify"`
}
