package notifier

type NotifierInterface interface {
	SendCodeBySms(phone string, code string) error
	SendLoginAndPasswordBySms(phone string, login string, password string) error
	SendEmailToOperator(subject, body string) error
}
