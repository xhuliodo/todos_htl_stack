package internal

type SignUpForm struct {
	Name          string
	NameError     string
	Email         string
	EmailError    string
	Password      string
	PasswordError string
	Error         string
	Message       string
	HasError      bool
}

type SignInForm struct {
	Email    string
	Password string
	Error    string
	HasError bool
	Message  string
}
