package system

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/system/application/command"
	"sipon-be/internal/modules/system/application/query"
	"sipon-be/internal/modules/system/infrastructure/identitygateway"
	"sipon-be/internal/modules/system/infrastructure/persistence"
	systemHTTP "sipon-be/internal/modules/system/interfaces/http"
)

// Module's exported surface is method-only, by design — zero exported fields.
// cmd/api/main.go gets RegisterRoutes(); other modules get Contract
// (contract.go) and nothing else. See docs/architecture/module-boundaries.md.
type Module struct {
	handler              *systemHTTP.SystemHandler
	authMiddleware       gin.HandlerFunc
	principalMiddleware  gin.HandlerFunc
	getUserScopeAccessUC *query.GetUserScopeAccessUseCase
	canAccessResourceUC  *query.CanAccessResourceUseCase
}

func NewModule(
	db *sql.DB,
	identityContract identity.Contract,
	authMiddleware gin.HandlerFunc,
	principalMiddleware gin.HandlerFunc,
) *Module {
	scopeRepo := persistence.NewPostgresScopeRepository(db)
	identityReader := identitygateway.New(identityContract)

	createScopeUC := command.NewCreateScopeUseCase(scopeRepo)
	updateScopeUC := command.NewUpdateScopeUseCase(scopeRepo)
	deleteScopeUC := command.NewDeleteScopeUseCase(scopeRepo)

	listScopesUC := query.NewListScopesUseCase(scopeRepo)
	getScopeUC := query.NewGetScopeUseCase(scopeRepo)
	getUserScopeAccessUC := query.NewGetUserScopeAccessUseCase(scopeRepo, identityReader)
	canAccessResourceUC := query.NewCanAccessResourceUseCase(scopeRepo, identityReader)

	handler := systemHTTP.NewSystemHandler(
		createScopeUC,
		updateScopeUC,
		deleteScopeUC,
		listScopesUC,
		getScopeUC,
	)

	return &Module{
		handler:              handler,
		authMiddleware:       authMiddleware,
		principalMiddleware:  principalMiddleware,
		getUserScopeAccessUC: getUserScopeAccessUC,
		canAccessResourceUC:  canAccessResourceUC,
	}
}

func (m *Module) RegisterRoutes(router gin.IRouter) {
	grp := router.Group("/")
	systemHTTP.RegisterRoutes(grp, m.handler, m.authMiddleware, m.principalMiddleware)
}

func (m *Module) GetUserScopeAccess(ctx context.Context, userID, scopeType string) (*UserScopeAccess, error) {
	res, err := m.getUserScopeAccessUC.Execute(ctx, userID, scopeType)
	if err != nil {
		return nil, err
	}
	return &UserScopeAccess{
		UserID:        res.UserID,
		ScopeType:     res.ScopeType,
		HasAccess:     res.HasAccess,
		HasFullAccess: res.HasFullAccess,
		AllowedCodes:  res.AllowedCodes,
	}, nil
}

func (m *Module) CanAccessResource(ctx context.Context, userID, scopeType string, resourceScopeCodes []string) (bool, error) {
	return m.canAccessResourceUC.Execute(ctx, userID, scopeType, resourceScopeCodes)
}
