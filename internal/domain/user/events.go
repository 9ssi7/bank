package user

type EventCreated struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	TempToken string `json:"temp_token"`
}
