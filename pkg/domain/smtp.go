package domain

type SMTPSettings struct {
	FromAddress           string
	FromName              string
	HELOName              string
	Host                  string
	Password              string
	Port                  int
	StartTLS              bool
	TLSInsecureSkipVerify bool
	Username              string
}
