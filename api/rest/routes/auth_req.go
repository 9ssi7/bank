package routes

type AuthLoginStartReq struct {
	Email string `json:"email" validate:"required,email"`
}

type AuthLoginReq struct {
	Code string `json:"code" validate:"required,numeric,len=4"`
}

type AuthRegisterReq struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type AuthRegistrationVerifyReq struct {
	Token string `params:"token" validate:"required,uuid"`
}
