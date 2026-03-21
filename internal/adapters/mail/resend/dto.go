package resend

type Request struct {
	From    string
	To      []string
	Subject string
	HTML    string
	Text    string
}

type Response struct {
	Id    string
	Error string
	Name  string
}
