package http

import (
	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/system/application/command"
	"sipon-be/internal/modules/system/application/dto"
	"sipon-be/internal/modules/system/application/query"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type SystemHandler struct {
	createScopeUC *command.CreateScopeUseCase
	updateScopeUC *command.UpdateScopeUseCase
	deleteScopeUC *command.DeleteScopeUseCase
	listScopesUC  *query.ListScopesUseCase
	getScopeUC    *query.GetScopeUseCase
}

func NewSystemHandler(
	createScopeUC *command.CreateScopeUseCase,
	updateScopeUC *command.UpdateScopeUseCase,
	deleteScopeUC *command.DeleteScopeUseCase,
	listScopesUC *query.ListScopesUseCase,
	getScopeUC *query.GetScopeUseCase,
) *SystemHandler {
	return &SystemHandler{
		createScopeUC: createScopeUC,
		updateScopeUC: updateScopeUC,
		deleteScopeUC: deleteScopeUC,
		listScopesUC:  listScopesUC,
		getScopeUC:    getScopeUC,
	}
}

func (h *SystemHandler) ListScopes(c *gin.Context) {
	var req dto.ListScopesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, err := h.listScopesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "scopes retrieved", items)
}

func (h *SystemHandler) GetScope(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getScopeUC.Execute(c.Request.Context(), id)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "scope retrieved", resp)
}

func (h *SystemHandler) CreateScope(c *gin.Context) {
	var req dto.CreateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createScopeUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "scope created", resp)
}

func (h *SystemHandler) UpdateScope(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateScopeUC.Execute(c.Request.Context(), id, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "scope updated", resp)
}

func (h *SystemHandler) DeleteScope(c *gin.Context) {
	id := c.Param("id")
	if err := h.deleteScopeUC.Execute(c.Request.Context(), id); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "scope deleted", nil)
}

func RequirePermission(permission string) gin.HandlerFunc {
	return middleware.RequirePermission(permission)
}
