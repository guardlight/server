package auth

type AuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
