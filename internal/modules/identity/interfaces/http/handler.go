package http

import (
	"strings"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/command"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/application/query"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/httperror"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/middleware"
	"sipon-be/internal/shared/respond"
)

type IdentityHandler struct {
	registerUC              *command.RegisterUseCase
	loginUC                 *command.LoginUseCase
	refreshTokenUC          *command.RefreshTokenUseCase
	changePasswordUC        *command.ChangePasswordLocalUseCase
	setPasswordUC           *command.SetPasswordLocalUseCase
	requestOTPUC            *command.RequestIdentityOTPUseCase
	verifyOTPUC             *command.VerifyIdentityOTPUseCase
	meUC                    *query.MeUseCase
	forgotPasswordUC        *command.ForgotPasswordUseCase
	resetPasswordUC         *command.ResetPasswordUseCase
	requestChangeIdentityUC *command.RequestChangeIdentityUseCase
	confirmChangeIdentityUC *command.ConfirmChangeIdentityUseCase
	getSessionUC            *query.GetSessionUseCase
	logoutUC                *command.LogoutUseCase
	tokenGen                application.TokenGenerator
	getProfileUC            *query.GetProfileUseCase
	updateProfileUC         *command.UpdateProfileUseCase
	checkUsernameUC         *query.CheckUsernameUseCase
	changeUsernameUC        *command.ChangeUsernameUseCase
	avatarPresignUC         *command.AvatarPresignUseCase
	avatarConfirmUC         *command.AvatarConfirmUseCase
	avatarDeleteUC          *command.AvatarDeleteUseCase

	createUserUC        *command.CreateUserUseCase
	resetUserPasswordUC *command.ResetUserPasswordUseCase
	deactivateUserUC    *command.DeactivateUserUseCase
	reactivateUserUC    *command.ReactivateUserUseCase
	listUsersUC         *query.ListUsersUseCase
	getUserUC           *query.GetUserUseCase

	listRolesUC          *query.ListRolesUseCase
	getRoleUC            *query.GetRoleUseCase
	listPermissionsUC    *query.ListPermissionsUseCase
	createRoleUC         *command.CreateRoleUseCase
	updateRoleUC         *command.UpdateRoleUseCase
	assignUserRoleUC     *command.AssignUserRoleUseCase
	updateUserRoleUC     *command.UpdateUserRoleUseCase
	deactivateUserRoleUC *command.DeactivateUserRoleUseCase
	reactivateUserRoleUC *command.ReactivateUserRoleUseCase
	deleteUserRoleUC     *command.DeleteUserRoleUseCase
	listUserRolesUC      *query.ListUserRolesUseCase
	getUserRoleUC        *query.GetUserRoleUseCase

	assignRolePermissionUC *command.AssignRolePermissionUseCase
	deleteRolePermissionUC *command.DeleteRolePermissionUseCase
	listScopesUC           *query.ListScopesUseCase
	assignRoleScopeUC      *command.AssignRoleScopeUseCase
	deleteRoleScopeUC      *command.DeleteRoleScopeUseCase
}

func NewIdentityHandler(
	registerUC *command.RegisterUseCase,
	loginUC *command.LoginUseCase,
	refreshTokenUC *command.RefreshTokenUseCase,
	changePasswordUC *command.ChangePasswordLocalUseCase,
	setPasswordUC *command.SetPasswordLocalUseCase,
	requestOTPUC *command.RequestIdentityOTPUseCase,
	verifyOTPUC *command.VerifyIdentityOTPUseCase,
	meUC *query.MeUseCase,
	forgotPasswordUC *command.ForgotPasswordUseCase,
	resetPasswordUC *command.ResetPasswordUseCase,
	requestChangeIdentityUC *command.RequestChangeIdentityUseCase,
	confirmChangeIdentityUC *command.ConfirmChangeIdentityUseCase,
	getSessionUC *query.GetSessionUseCase,
	logoutUC *command.LogoutUseCase,
	getProfileUC *query.GetProfileUseCase,
	updateProfileUC *command.UpdateProfileUseCase,
	checkUsernameUC *query.CheckUsernameUseCase,
	changeUsernameUC *command.ChangeUsernameUseCase,
	avatarPresignUC *command.AvatarPresignUseCase,
	avatarConfirmUC *command.AvatarConfirmUseCase,
	avatarDeleteUC *command.AvatarDeleteUseCase,
	createUserUC *command.CreateUserUseCase,
	resetUserPasswordUC *command.ResetUserPasswordUseCase,
	deactivateUserUC *command.DeactivateUserUseCase,
	reactivateUserUC *command.ReactivateUserUseCase,
	listUsersUC *query.ListUsersUseCase,
	getUserUC *query.GetUserUseCase,
	listRolesUC *query.ListRolesUseCase,
	getRoleUC *query.GetRoleUseCase,
	listPermissionsUC *query.ListPermissionsUseCase,
	createRoleUC *command.CreateRoleUseCase,
	updateRoleUC *command.UpdateRoleUseCase,
	assignUserRoleUC *command.AssignUserRoleUseCase,
	updateUserRoleUC *command.UpdateUserRoleUseCase,
	deactivateUserRoleUC *command.DeactivateUserRoleUseCase,
	reactivateUserRoleUC *command.ReactivateUserRoleUseCase,
	deleteUserRoleUC *command.DeleteUserRoleUseCase,
	listUserRolesUC *query.ListUserRolesUseCase,
	getUserRoleUC *query.GetUserRoleUseCase,
	assignRolePermissionUC *command.AssignRolePermissionUseCase,
	deleteRolePermissionUC *command.DeleteRolePermissionUseCase,
	listScopesUC *query.ListScopesUseCase,
	assignRoleScopeUC *command.AssignRoleScopeUseCase,
	deleteRoleScopeUC *command.DeleteRoleScopeUseCase,
	tokenGen application.TokenGenerator,
) *IdentityHandler {
	return &IdentityHandler{
		registerUC:              registerUC,
		loginUC:                 loginUC,
		refreshTokenUC:          refreshTokenUC,
		changePasswordUC:        changePasswordUC,
		setPasswordUC:           setPasswordUC,
		requestOTPUC:            requestOTPUC,
		verifyOTPUC:             verifyOTPUC,
		meUC:                    meUC,
		forgotPasswordUC:        forgotPasswordUC,
		resetPasswordUC:         resetPasswordUC,
		requestChangeIdentityUC: requestChangeIdentityUC,
		confirmChangeIdentityUC: confirmChangeIdentityUC,
		getSessionUC:            getSessionUC,
		logoutUC:                logoutUC,
		tokenGen:                tokenGen,
		getProfileUC:            getProfileUC,
		updateProfileUC:         updateProfileUC,
		checkUsernameUC:         checkUsernameUC,
		changeUsernameUC:        changeUsernameUC,
		avatarPresignUC:         avatarPresignUC,
		avatarConfirmUC:         avatarConfirmUC,
		avatarDeleteUC:          avatarDeleteUC,
		createUserUC:            createUserUC,
		resetUserPasswordUC:     resetUserPasswordUC,
		deactivateUserUC:        deactivateUserUC,
		reactivateUserUC:        reactivateUserUC,
		listUsersUC:             listUsersUC,
		getUserUC:               getUserUC,
		listRolesUC:             listRolesUC,
		getRoleUC:               getRoleUC,
		listPermissionsUC:       listPermissionsUC,
		createRoleUC:            createRoleUC,
		updateRoleUC:            updateRoleUC,
		assignUserRoleUC:        assignUserRoleUC,
		updateUserRoleUC:        updateUserRoleUC,
		deactivateUserRoleUC:    deactivateUserRoleUC,
		reactivateUserRoleUC:    reactivateUserRoleUC,
		deleteUserRoleUC:        deleteUserRoleUC,
		listUserRolesUC:         listUserRolesUC,
		getUserRoleUC:           getUserRoleUC,
		assignRolePermissionUC:  assignRolePermissionUC,
		deleteRolePermissionUC:  deleteRolePermissionUC,
		listScopesUC:            listScopesUC,
		assignRoleScopeUC:       assignRoleScopeUC,
		deleteRoleScopeUC:       deleteRoleScopeUC,
	}
}

func (h *IdentityHandler) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func (h *IdentityHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.registerUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "register success", resp)
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.loginUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "login success", resp)
}

func (h *IdentityHandler) RequestIdentityOTP(c *gin.Context) {
	var req dto.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.requestOTPUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) VerifyIdentityOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.verifyOTPUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.refreshTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "refresh token success", resp)
}

func (h *IdentityHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.forgotPasswordUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.resetPasswordUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) GetSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getSessionUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "session retrieved", resp)
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	token := h.extractToken(c)
	if token == "" {
		httperror.Handle(c, kernel.New(application.ErrCodeUnauthorized))
		return
	}
	claims, err := h.tokenGen.ParseAccessToken(token)
	if err != nil {
		httperror.Handle(c, kernel.Wrap(application.ErrCodeUnauthorized, err))
		return
	}
	if err := h.logoutUC.Execute(c.Request.Context(), claims.SessionID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "Logged out successfully", nil)
}

func (h *IdentityHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.meUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "me retrieved", resp)
}

func (h *IdentityHandler) Profile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getProfileUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "profile retrieved", resp)
}

func (h *IdentityHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.changePasswordUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) SetPassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.setPasswordUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) RequestChangeIdentityEmail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.RequestChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.requestChangeIdentityUC.Execute(c.Request.Context(), userID, domain.LoginIdentifierKindEmail, req.NewEmail)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) ConfirmChangeIdentityEmail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ConfirmChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.confirmChangeIdentityUC.Execute(c.Request.Context(), userID, domain.LoginIdentifierKindEmail, req.OTP)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) RequestChangeIdentityPhone(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.RequestChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.requestChangeIdentityUC.Execute(c.Request.Context(), userID, domain.LoginIdentifierKindPhone, req.NewPhone)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) ConfirmChangeIdentityPhone(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ConfirmChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.confirmChangeIdentityUC.Execute(c.Request.Context(), userID, domain.LoginIdentifierKindPhone, req.OTP)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateProfileUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) CheckUsername(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := c.Query("username")
	if username == "" {
		httperror.Handle(c, kernel.New(application.ErrCodeUnprocessableEntity))
		return
	}
	resp, err := h.checkUsernameUC.Execute(c.Request.Context(), userID, username)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "check username", resp)
}

func (h *IdentityHandler) ChangeUsername(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.changeUsernameUC.Execute(c.Request.Context(), userID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "change username success", resp)
}

func (h *IdentityHandler) AvatarPresign(c *gin.Context) {
	var req dto.AvatarPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.avatarPresignUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "avatar presign url generated", resp)
}

func (h *IdentityHandler) AvatarConfirm(c *gin.Context) {
	userID := middleware.GetUserID(c)
	key := c.Query("key")
	if key == "" {
		httperror.Handle(c, kernel.New(application.ErrCodeUnprocessableEntity))
		return
	}
	resp, err := h.avatarConfirmUC.Execute(c.Request.Context(), userID, key)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "Avatar confirmed successfully", resp)
}

func (h *IdentityHandler) AvatarDelete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.avatarDeleteUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, resp.Message, nil)
}

func (h *IdentityHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listUsersUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "users retrieved", items, meta)
}

func (h *IdentityHandler) GetUser(c *gin.Context) {
	userID := c.Param("user_id")
	resp, err := h.getUserUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user retrieved", resp)
}

func (h *IdentityHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createUserUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "User created successfully", resp)
}

func (h *IdentityHandler) ResetUserPassword(c *gin.Context) {
	userID := c.Param("user_id")
	resp, err := h.resetUserPasswordUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User password reset successfully", resp)
}

func (h *IdentityHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("user_id")
	resp, err := h.deactivateUserUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User deactivated successfully", resp)
}

func (h *IdentityHandler) ReactivateUser(c *gin.Context) {
	userID := c.Param("user_id")
	resp, err := h.reactivateUserUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User reactivated successfully", resp)
}

func (h *IdentityHandler) ListRoles(c *gin.Context) {
	var req dto.ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listRolesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "roles retrieved", items, meta)
}

func (h *IdentityHandler) GetRole(c *gin.Context) {
	roleID := c.Param("role_id")
	resp, err := h.getRoleUC.Execute(c.Request.Context(), roleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "role retrieved", resp)
}

func (h *IdentityHandler) ListPermissions(c *gin.Context) {
	items := h.listPermissionsUC.Execute(c.Request.Context())
	respond.OK(c, "permissions retrieved", items)
}

func (h *IdentityHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.createRoleUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "role created", resp)
}

func (h *IdentityHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("role_id")
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateRoleUC.Execute(c.Request.Context(), roleID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "role updated", resp)
}

func (h *IdentityHandler) AssignRolePermission(c *gin.Context) {
	roleID := c.Param("role_id")
	assignedBy := middleware.GetUserID(c)
	var req dto.AssignRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.assignRolePermissionUC.Execute(c.Request.Context(), roleID, assignedBy, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "Permission assigned to role successfully", resp)
}

func (h *IdentityHandler) DeleteRolePermission(c *gin.Context) {
	roleID := c.Param("role_id")
	permissionKey := c.Param("permission_key")
	resp, err := h.deleteRolePermissionUC.Execute(c.Request.Context(), roleID, permissionKey)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "Permission removed from role successfully", resp)
}

func (h *IdentityHandler) ListUserRoles(c *gin.Context) {
	var req dto.ListUserRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	items, meta, err := h.listUserRolesUC.Execute(c.Request.Context(), req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.SuccessWithMeta(c, 200, "user roles retrieved", items, meta)
}

func (h *IdentityHandler) GetUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	resp, err := h.getUserRoleUC.Execute(c.Request.Context(), userRoleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "user role retrieved", resp)
}

func (h *IdentityHandler) AssignUserRole(c *gin.Context) {
	assignedBy := middleware.GetUserID(c)
	var req dto.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.assignUserRoleUC.Execute(c.Request.Context(), assignedBy, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "Role assigned to user successfully", resp)
}

func (h *IdentityHandler) UpdateUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.updateUserRoleUC.Execute(c.Request.Context(), userRoleID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User role updated successfully", resp)
}

func (h *IdentityHandler) DeactivateUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	resp, err := h.deactivateUserRoleUC.Execute(c.Request.Context(), userRoleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User role deactivated successfully", resp)
}

func (h *IdentityHandler) ReactivateUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	resp, err := h.reactivateUserRoleUC.Execute(c.Request.Context(), userRoleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User role reactivated successfully", resp)
}

func (h *IdentityHandler) DeleteUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	if err := h.deleteUserRoleUC.Execute(c.Request.Context(), userRoleID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "User role deleted successfully", nil)
}

func (h *IdentityHandler) ListScopes(c *gin.Context) {
	roleID := c.Param("role_id")
	items, err := h.listScopesUC.Execute(c.Request.Context(), roleID)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "scopes retrieved", items)
}

func (h *IdentityHandler) AssignRoleScope(c *gin.Context) {
	roleID := c.Param("role_id")
	var req dto.AssignScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.Handle(c, err)
		return
	}
	resp, err := h.assignRoleScopeUC.Execute(c.Request.Context(), roleID, req)
	if err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.Created(c, "Scope assigned to role successfully", resp)
}

func (h *IdentityHandler) DeleteRoleScope(c *gin.Context) {
	scopeID := c.Param("scope_id")
	if err := h.deleteRoleScopeUC.Execute(c.Request.Context(), scopeID); err != nil {
		httperror.Handle(c, err)
		return
	}
	respond.OK(c, "Scope removed from role successfully", nil)
}
