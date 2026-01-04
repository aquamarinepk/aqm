package auth

import (
	"time"

	"github.com/aquamarinepk/aqm/crypto"
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id" db:"id" bson:"_id"`
	Username string    `json:"username" db:"username" bson:"username"`
	Name     string    `json:"name" db:"name" bson:"name"`

	EmailCT     []byte `json:"-" db:"email_ct" bson:"email_ct"`
	EmailIV     []byte `json:"-" db:"email_iv" bson:"email_iv"`
	EmailTag    []byte `json:"-" db:"email_tag" bson:"email_tag"`
	EmailLookup []byte `json:"-" db:"email_lookup" bson:"email_lookup"`

	PasswordHash []byte `json:"-" db:"password_hash" bson:"pass_hash"`
	PasswordSalt []byte `json:"-" db:"password_salt" bson:"pass_salt"`

	MFASecretCT []byte `json:"-" db:"mfa_secret_ct" bson:"mfa_secret_ct,omitempty"`

	PINCT     []byte `json:"-" db:"pin_ct" bson:"pin_ct,omitempty"`
	PINIV     []byte `json:"-" db:"pin_iv" bson:"pin_iv,omitempty"`
	PINTag    []byte `json:"-" db:"pin_tag" bson:"pin_tag,omitempty"`
	PINLookup []byte `json:"-" db:"pin_lookup" bson:"pin_lookup,omitempty"`

	Status UserStatus `json:"status" db:"status" bson:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	CreatedBy string    `json:"created_by" db:"created_by" bson:"created_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
	UpdatedBy string    `json:"updated_by" db:"updated_by" bson:"updated_by"`
}

func NewUser() *User {
	return &User{
		Status: UserStatusActive,
	}
}

func (u *User) EnsureID() {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
}

func (u *User) BeforeCreate() {
	u.EnsureID()
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	u.Username = NormalizeUsername(u.Username)
	u.Name = NormalizeDisplayName(u.Name)
}

func (u *User) BeforeUpdate() {
	u.UpdatedAt = time.Now()
	u.Username = NormalizeUsername(u.Username)
	u.Name = NormalizeDisplayName(u.Name)
}

func (u *User) SetEmail(email string, encryptionKey, signingKey []byte) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	normalized := NormalizeEmail(email)

	ct, iv, tag, err := crypto.EncryptEmail(normalized, encryptionKey)
	if err != nil {
		return ErrEncryptionFailed
	}

	lookup := crypto.ComputeLookupHash(normalized, signingKey)

	u.EmailCT = []byte(ct)
	u.EmailIV = []byte(iv)
	u.EmailTag = []byte(tag)
	u.EmailLookup = []byte(lookup)

	return nil
}

func (u *User) GetEmail(encryptionKey []byte) (string, error) {
	email, err := crypto.DecryptEmail(
		string(u.EmailCT),
		string(u.EmailIV),
		string(u.EmailTag),
		encryptionKey,
	)
	if err != nil {
		return "", ErrDecryptionFailed
	}
	return email, nil
}

func (u *User) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return ErrPasswordHashFailed
	}

	hash := crypto.HashPassword(password, salt)
	if hash == nil {
		return ErrPasswordHashFailed
	}

	u.PasswordHash = hash
	u.PasswordSalt = salt

	return nil
}

func (u *User) VerifyPassword(password string) bool {
	return crypto.VerifyPassword(password, u.PasswordHash, u.PasswordSalt)
}

func (u *User) SetPIN(pin string, encryptionKey, signingKey []byte) error {
	if len(pin) < 4 || len(pin) > 8 {
		return ErrInvalidPassword
	}

	ct, iv, tag, err := crypto.EncryptEmail(pin, encryptionKey)
	if err != nil {
		return ErrEncryptionFailed
	}

	lookup := crypto.ComputeLookupHash(pin, signingKey)

	u.PINCT = []byte(ct)
	u.PINIV = []byte(iv)
	u.PINTag = []byte(tag)
	u.PINLookup = []byte(lookup)

	return nil
}

func (u *User) VerifyPIN(pin string, signingKey []byte) bool {
	if len(u.PINLookup) == 0 {
		return false
	}

	lookup := crypto.ComputeLookupHash(pin, signingKey)
	return string(u.PINLookup) == lookup
}

func (u *User) Validate() error {
	if err := ValidateUsername(u.Username); err != nil {
		return err
	}
	if err := ValidateDisplayName(u.Name); err != nil {
		return err
	}
	if !u.Status.IsValid() {
		return ErrInactiveAccount
	}
	return nil
}
