package gateway

import "context"

const ScopePersonsDelete = "persons:delete"

// AdminRecognitionGateway executa operações administrativas no serviço de reconhecimento
// facial usando o token informado pelo administrador (não o token do servidor).
// Implementação: internal/framework/booprecognition.
type AdminRecognitionGateway interface {
	// ValidateToken retorna entity.ErrInvalidToken se o parceiro recusar o token.
	ValidateToken(ctx context.Context, token string) (scopes []string, err error)
	// DeletePerson remove a pessoa no parceiro. Pessoa inexistente lá não é erro.
	DeletePerson(ctx context.Context, token string, partnerPersonID int64) error
}
