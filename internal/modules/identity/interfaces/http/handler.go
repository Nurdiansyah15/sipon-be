package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/command"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/application/query"
	domainErrors "sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/middleware"
)

type IdentityHandler struct {
	registerUC             *command.RegisterUseCase
	loginUC                *command.LoginUseCase
	refreshTokenUC         *command.RefreshTokenUseCase
	changePasswordUC       *command.ChangePasswordLocalUseCase
	setPasswordUC          *command.SetPasswordLocalUseCase
	requestOTPUC           *command.RequestIdentityOTPUseCase
	verifyOTPUC            *command.VerifyIdentityOTPUseCase
	meUC                   *query.MeUseCase
	forgotPasswordUC       *command.ForgotPasswordUseCase
	resetPasswordUC        *command.ResetPasswordUseCase
	requestChangeIdentityUC *command.RequestChangeIdentityUseCase
	confirmChangeIdentityUC *command.ConfirmChangeIdentityUseCase
	getSessionUC           *query.GetSessionUseCase
	logoutUC               *command.LogoutUseCase
	getProfileUC           *query.GetProfileUseCase
	updateProfileUC        *command.UpdateProfileUseCase
	checkUsernameUC        *query.CheckUsernameUseCase
	changeUsernameUC       *command.ChangeUsernameUseCase
	avatarPresignUC        *command.AvatarPresignUseCase
	avatarConfirmUC        *command.AvatarConfirmUseCase
	avatarDeleteUC         *command.AvatarDeleteUseCase

	createUserUC         *command.CreateUserUseCase
	resetUserPasswordUC  *command.ResetUserPasswordUseCase
	deactivateUserUC     *command.DeactivateUserUseCase
	reactivateUserUC     *command.ReactivateUserUseCase
	listUsersUC          *query.ListUsersUseCase
	getUserUC            *query.GetUserUseCase

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
	assignRolePermissionUC *command.AssignRolePermissionUseCase,
	deleteRolePermissionUC *command.DeleteRolePermissionUseCase,
	listScopesUC *query.ListScopesUseCase,
	assignRoleScopeUC *command.AssignRoleScopeUseCase,
	deleteRoleScopeUC *command.DeleteRoleScopeUseCase,
) *IdentityHandler {
	return &IdentityHandler{
		registerUC:             registerUC,
		loginUC:                loginUC,
		refreshTokenUC:         refreshTokenUC,
		changePasswordUC:       changePasswordUC,
		setPasswordUC:          setPasswordUC,
		requestOTPUC:           requestOTPUC,
		verifyOTPUC:            verifyOTPUC,
		meUC:                   meUC,
		forgotPasswordUC:       forgotPasswordUC,
		resetPasswordUC:        resetPasswordUC,
		requestChangeIdentityUC: requestChangeIdentityUC,
		confirmChangeIdentityUC: confirmChangeIdentityUC,
		getSessionUC:           getSessionUC,
		logoutUC:               logoutUC,
		getProfileUC:           getProfileUC,
		updateProfileUC:        updateProfileUC,
		checkUsernameUC:        checkUsernameUC,
		changeUsernameUC:       changeUsernameUC,
		avatarPresignUC:        avatarPresignUC,
		avatarConfirmUC:        avatarConfirmUC,
		avatarDeleteUC:         avatarDeleteUC,
		createUserUC:           createUserUC,
		resetUserPasswordUC:    resetUserPasswordUC,
		deactivateUserUC:       deactivateUserUC,
		reactivateUserUC:       reactivateUserUC,
		listUsersUC:            listUsersUC,
		getUserUC:              getUserUC,
		listRolesUC:            listRolesUC,
		getRoleUC:              getRoleUC,
		listPermissionsUC:      listPermissionsUC,
		createRoleUC:           createRoleUC,
		updateRoleUC:           updateRoleUC,
		assignUserRoleUC:       assignUserRoleUC,
		updateUserRoleUC:       updateUserRoleUC,
		deactivateUserRoleUC:   deactivateUserRoleUC,
		reactivateUserRoleUC:   reactivateUserRoleUC,
		deleteUserRoleUC:       deleteUserRoleUC,
		listUserRolesUC:        listUserRolesUC,
		assignRolePermissionUC: assignRolePermissionUC,
		deleteRolePermissionUC: deleteRolePermissionUC,
		listScopesUC:           listScopesUC,
		assignRoleScopeUC:      assignRoleScopeUC,
		deleteRoleScopeUC:      deleteRoleScopeUC,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.registerUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.loginUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) RequestIdentityOTP(c *gin.Context) {
	var req dto.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.requestOTPUC.Execute(c.Request.Context(), req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

func (h *IdentityHandler) VerifyIdentityOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.verifyOTPUC.Execute(c.Request.Context(), req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Identity verified successfully"})
}

func (h *IdentityHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.refreshTokenUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.forgotPasswordUC.Execute(c.Request.Context(), req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset OTP sent if email exists"})
}

func (h *IdentityHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.resetPasswordUC.Execute(c.Request.Context(), req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *IdentityHandler) GetSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getSessionUC.Execute(c.Request.Context(), userID)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	token := h.extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": "No valid access token found"}})
		return
	}
	if err := h.logoutUC.Execute(c.Request.Context(), token); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *IdentityHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.meUC.Execute(c.Request.Context(), userID)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) Profile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	resp, err := h.getProfileUC.Execute(c.Request.Context(), userID)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.changePasswordUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *IdentityHandler) SetPassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.SetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.setPasswordUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password set successfully"})
}

func (h *IdentityHandler) RequestChangeIdentityEmail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangeIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.requestChangeIdentityUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to new identity"})
}

func (h *IdentityHandler) ConfirmChangeIdentityEmail(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangeIdentityConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.confirmChangeIdentityUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Identity changed successfully"})
}

func (h *IdentityHandler) RequestChangeIdentityPhone(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangeIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.requestChangeIdentityUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to new identity"})
}

func (h *IdentityHandler) ConfirmChangeIdentityPhone(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangeIdentityConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.confirmChangeIdentityUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Identity changed successfully"})
}

func (h *IdentityHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.updateProfileUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *IdentityHandler) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "username query parameter is required"}})
		return
	}
	resp, err := h.checkUsernameUC.Execute(c.Request.Context(), username)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) ChangeUsername(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.ChangeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.changeUsernameUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Username changed successfully"})
}

func (h *IdentityHandler) AvatarPresign(c *gin.Context) {
	var req dto.AvatarPresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.avatarPresignUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) AvatarConfirm(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req dto.AvatarConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.avatarConfirmUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Avatar confirmed successfully"})
}

func (h *IdentityHandler) AvatarDelete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.avatarDeleteUC.Execute(c.Request.Context(), userID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Avatar deleted successfully"})
}

func (h *IdentityHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.listUsersUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) GetUser(c *gin.Context) {
	userID := c.Param("user_id")
	resp, err := h.getUserUC.Execute(c.Request.Context(), userID)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) CreateUser(c *gin.Context) {
	createdBy := middleware.GetUserID(c)
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.createUserUC.Execute(c.Request.Context(), req, createdBy); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *IdentityHandler) ResetUserPassword(c *gin.Context) {
	userID := c.Param("user_id")
	var req dto.ResetUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.resetUserPasswordUC.Execute(c.Request.Context(), userID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User password reset successfully"})
}

func (h *IdentityHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("user_id")
	if err := h.deactivateUserUC.Execute(c.Request.Context(), userID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

func (h *IdentityHandler) ReactivateUser(c *gin.Context) {
	userID := c.Param("user_id")
	if err := h.reactivateUserUC.Execute(c.Request.Context(), userID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User reactivated successfully"})
}

func (h *IdentityHandler) ListRoles(c *gin.Context) {
	var req dto.ListRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.listRolesUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) GetRole(c *gin.Context) {
	roleID := c.Param("role_id")
	resp, err := h.getRoleUC.Execute(c.Request.Context(), roleID)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) ListPermissions(c *gin.Context) {
	items, err := h.listPermissionsUC.Execute(c.Request.Context())
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"permissions": items}})
}

func (h *IdentityHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.createRoleUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (h *IdentityHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("role_id")
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.updateRoleUC.Execute(c.Request.Context(), roleID, req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) AssignRolePermission(c *gin.Context) {
	roleID := c.Param("role_id")
	assignedBy := middleware.GetUserID(c)
	var req dto.AssignRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.assignRolePermissionUC.Execute(c.Request.Context(), roleID, assignedBy, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permission assigned to role successfully"})
}

func (h *IdentityHandler) DeleteRolePermission(c *gin.Context) {
	roleID := c.Param("role_id")
	permissionKey := c.Param("permission_key")
	if err := h.deleteRolePermissionUC.Execute(c.Request.Context(), roleID, permissionKey); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permission removed from role successfully"})
}

func (h *IdentityHandler) ListUserRoles(c *gin.Context) {
	var req dto.ListUserRolesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	resp, err := h.listUserRolesUC.Execute(c.Request.Context(), req)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) AssignUserRole(c *gin.Context) {
	assignedBy := middleware.GetUserID(c)
	var req dto.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.assignUserRoleUC.Execute(c.Request.Context(), assignedBy, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Role assigned to user successfully"})
}

func (h *IdentityHandler) UpdateUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.updateUserRoleUC.Execute(c.Request.Context(), userRoleID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

func (h *IdentityHandler) DeactivateUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	if err := h.deactivateUserRoleUC.Execute(c.Request.Context(), userRoleID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role deactivated successfully"})
}

func (h *IdentityHandler) ReactivateUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	if err := h.reactivateUserRoleUC.Execute(c.Request.Context(), userRoleID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role reactivated successfully"})
}

func (h *IdentityHandler) DeleteUserRole(c *gin.Context) {
	userRoleID := c.Param("user_role_id")
	if err := h.deleteUserRoleUC.Execute(c.Request.Context(), userRoleID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role deleted successfully"})
}

func (h *IdentityHandler) ListScopes(c *gin.Context) {
	roleID := c.Param("role_id")
	resp, err := h.listScopesUC.Execute(c.Request.Context(), roleID)
	if err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *IdentityHandler) AssignRoleScope(c *gin.Context) {
	roleID := c.Param("role_id")
	var req dto.AssignScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	if err := h.assignRoleScopeUC.Execute(c.Request.Context(), roleID, req); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Scope assigned to role successfully"})
}

func (h *IdentityHandler) DeleteRoleScope(c *gin.Context) {
	scopeID := c.Param("scope_id")
	if err := h.deleteRoleScopeUC.Execute(c.Request.Context(), scopeID); err != nil {
		status, code, msg := mapError(err)
		c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Scope removed from role successfully"})
}

func mapError(err error) (int, string, string) {
	if err == nil {
		return http.StatusInternalServerError, "INTERNAL_ERROR", "Unknown error"
	}

	code := string(kernel.Code("INTERNAL_ERROR"))
	message := err.Error()

	for cur := err; cur != nil; {
		if ae, ok := cur.(*kernel.AppError); ok {
			code = string(ae.Code)
			if ae.Err != nil {
				message = ae.Err.Error()
			}
			break
		}
		inner := cur
		if wrapped, ok := cur.(interface{ Unwrap() error }); ok {
			inner = wrapped.Unwrap()
			if inner == cur {
				break
			}
		}
		cur = inner
	}

	switch kernel.Code(code) {
	case application.ErrCodeUserNotFound:
		return http.StatusNotFound, code, "User not found"
	case application.ErrCodeDuplicateEmail, application.ErrCodeDuplicatePhone, application.ErrCodeDuplicateUsername:
		return http.StatusConflict, code, message
	case application.ErrCodeInvalidCredentials:
		return http.StatusUnauthorized, code, "Invalid credentials"
	case application.ErrCodeInvalidToken, application.ErrCodeTokenExpired:
		return http.StatusUnauthorized, code, "Invalid or expired token"
	case application.ErrCodeSessionRevoked:
		return http.StatusUnauthorized, code, "Session has been revoked"
	case application.ErrCodeRateLimited:
		return http.StatusTooManyRequests, code, "Rate limit exceeded"

	case domainErrors.ErrCodeRoleNotFound:
		return http.StatusNotFound, code, "Role not found"
	case domainErrors.ErrCodeRoleNotAssignable:
		return http.StatusForbidden, code, message
	case domainErrors.ErrCodeRoleCannotDeleteSystem:
		return http.StatusForbidden, code, message
	case domainErrors.ErrCodeRoleScopeMismatch:
		return http.StatusBadRequest, code, message
	case domainErrors.ErrCodeUserRoleAlreadyAssigned:
		return http.StatusConflict, code, message
	case domainErrors.ErrCodeUserRoleExpired, domainErrors.ErrCodeUserRoleNotActive:
		return http.StatusBadRequest, code, message
	case domainErrors.ErrCodeInvalidPermissionKey, domainErrors.ErrCodeInvalidScopeType, domainErrors.ErrCodeInvalidRoleType:
		return http.StatusBadRequest, code, message
	case domainErrors.ErrCodeVerificationCodeNotFound:
		return http.StatusNotFound, code, "Verification code not found"
	case domainErrors.ErrCodeVerificationCodeExpired:
		return http.StatusGone, code, "Verification code expired"
	case domainErrors.ErrCodeVerificationCodeUsed:
		return http.StatusConflict, code, "Verification code already used"
	case domainErrors.ErrCodeVerificationCodeMismatch:
		return http.StatusBadRequest, code, "Verification code mismatch"
	case domainErrors.ErrCodeVerificationInvalidPurpose, domainErrors.ErrCodeVerificationNewIdentityEmpty:
		return http.StatusBadRequest, code, message
	case domainErrors.ErrCodeTooManyVerificationCode:
		return http.StatusTooManyRequests, code, message
	case domainErrors.ErrCodeIdentityNotVerified:
		return http.StatusForbidden, code, message

	case domainErrors.ErrCodeUserBanned:
		return http.StatusForbidden, code, "User account is banned"
	case domainErrors.ErrCodeUserLockedOut:
		return http.StatusTooManyRequests, code, "Account temporarily locked"
	case domainErrors.ErrCodeUserNotActive:
		return http.StatusForbidden, code, "User account is not active"
	case domainErrors.ErrCodeCredentialNotLocal:
		return http.StatusBadRequest, code, message
	case domainErrors.ErrCodeUsernameAlreadySet:
		return http.StatusConflict, code, message
	case domainErrors.ErrCodeNoPrimaryIdentity:
		return http.StatusBadRequest, code, message
	case domainErrors.ErrCodeUserAlreadyActive, domainErrors.ErrCodeUserAlreadyBanned, domainErrors.ErrCodeUserAlreadyDeleted:
		return http.StatusConflict, code, message
	}

	return http.StatusInternalServerError, "INTERNAL_ERROR", message
}
