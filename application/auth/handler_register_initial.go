package auth

import (
	"time"

	"github.com/boreq/velo/domain/auth"
	"github.com/pkg/errors"
)

type RegisterInitial struct {
	Username auth.ValidatedUsername
	Password auth.ValidatedPassword
}

type RegisterInitialHandler struct {
	passwordHasher      PasswordHasher
	transactionProvider TransactionProvider
	uuidGenerator       UUIDGenerator
}

func NewRegisterInitialHandler(
	passwordHasher PasswordHasher,
	transactionProvider TransactionProvider,
	uuidGenerator UUIDGenerator,
) *RegisterInitialHandler {
	return &RegisterInitialHandler{
		passwordHasher:      passwordHasher,
		transactionProvider: transactionProvider,
		uuidGenerator:       uuidGenerator,
	}
}

func (h *RegisterInitialHandler) Execute(cmd RegisterInitial) error {
	if cmd.Username.IsZero() {
		return errors.New("zero value of username")
	}

	if cmd.Password.IsZero() {
		return errors.New("zero value of password")
	}

	passwordHash, err := h.passwordHasher.Hash(cmd.Password.String())
	if err != nil {
		return errors.Wrap(err, "hashing the password failed")
	}

	uuid, err := h.uuidGenerator.Generate()
	if err != nil {
		return errors.Wrap(err, "could not generate an uuid")
	}

	userUUID, err := auth.NewUserUUID(uuid)
	if err != nil {
		return errors.Wrap(err, "could not create a user uuid")
	}

	u := User{
		UUID:          userUUID,
		Username:      cmd.Username.String(),
		Password:      passwordHash,
		Administrator: true,
		Created:       time.Now(),
		LastSeen:      time.Now(),
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		n, err := r.Users.Count()
		if err != nil {
			return errors.Wrap(err, "could not get a count")
		}
		if n != 0 {
			return errors.New("there are existing users")
		}
		return r.Users.Put(u)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
