package entity

import (
	"net/mail"
	"regexp"
	"strings"
)

const DefaultCode = "0000"

var codeRe = regexp.MustCompile(`^[0-9]{4}$`)

type Person struct {
	ID              int64
	Name            string
	Email           string
	PartnerPersonID int64 // identification no serviço de reconhecimento facial
	ImagePath       string
	CodeHash        []byte
	CreatedAt       string
}

// NormalizePersonInput valida e normaliza os dados de cadastro.
func NormalizePersonInput(name, email, code string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return "", "", "", ValidationError("name é obrigatório (máx. 120 caracteres)")
	}
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", "", "", ValidationError("email inválido")
	}
	if code == "" {
		code = DefaultCode
	}
	if err := ValidateCode(code); err != nil {
		return "", "", "", err
	}
	return name, strings.ToLower(addr.Address), code, nil
}

func ValidateCode(code string) error {
	if !codeRe.MatchString(code) {
		return ValidationError("code deve ter exatamente 4 dígitos")
	}
	return nil
}

// MaskedEmail devolve o e-mail sem expor o usuário completo.
func (p *Person) MaskedEmail() string {
	user, domain, ok := strings.Cut(p.Email, "@")
	if !ok || user == "" {
		return "***"
	}
	return user[:1] + "***@" + domain
}
