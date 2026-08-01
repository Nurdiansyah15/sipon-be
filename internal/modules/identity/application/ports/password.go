package ports

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hashed, plain string) error
}
