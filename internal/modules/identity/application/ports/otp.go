package ports

type OTPGenerator interface {
	Generate() (string, error)
}
