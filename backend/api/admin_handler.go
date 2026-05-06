package api

import (
	"strconv"

	"backend/model"
	"backend/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc   *service.AdminService
	aiSvc *service.AIService
}

func NewAdminHandler(svc *service.AdminService, aiSvc *service.AIService) *AdminHandler {
	return &AdminHandler{svc: svc, aiSvc: aiSvc}
}

func (h *AdminHandler) HandleListUsers(c *gin.Context) {
	list, err := h.svc.ListUsers(c.Query("keyword"), c.Query("role"))
	if err != nil {
		SendError(c, "500", "获取用户列表失败: "+err.Error())
		return
	}
	SendSuccess(c, list)
}

func (h *AdminHandler) HandleCreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateUser(req); err != nil {
		SendError(c, "500", "创建用户失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleUpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		SendError(c, "400", "用户 ID 不合法")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateUser(uint(id), req); err != nil {
		SendError(c, "500", "更新用户失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleDeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		SendError(c, "400", "用户 ID 不合法")
		return
	}
	if err := h.svc.DeleteUser(uint(id), GetUserID(c)); err != nil {
		SendError(c, "500", "删除用户失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		SendError(c, "500", "获取角色列表失败: "+err.Error())
		return
	}
	SendSuccess(c, roles)
}

func (h *AdminHandler) HandleCreateRole(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateRole(role); err != nil {
		SendError(c, "500", "创建角色失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleDeleteRole(c *gin.Context) {
	roleCode := c.Param("roleCode")
	if err := h.svc.DeleteRole(roleCode); err != nil {
		SendError(c, "500", "删除角色失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleListPermissions(c *gin.Context) {
	permissions, err := h.svc.ListPermissions()
	if err != nil {
		SendError(c, "500", "获取权限列表失败: "+err.Error())
		return
	}
	SendSuccess(c, permissions)
}

func (h *AdminHandler) HandleCreatePermission(c *gin.Context) {
	var p model.Permission
	if err := c.ShouldBindJSON(&p); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreatePermission(p); err != nil {
		SendError(c, "500", "创建权限点失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleUpdatePermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		SendError(c, "400", "权限 ID 不合法")
		return
	}

	var p model.Permission
	if err := c.ShouldBindJSON(&p); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdatePermission(uint(id), p); err != nil {
		SendError(c, "500", "更新权限点失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleDeletePermission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		SendError(c, "400", "权限 ID 不合法")
		return
	}
	if err := h.svc.DeletePermission(uint(id)); err != nil {
		SendError(c, "500", "删除权限点失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleGetRolePermissions(c *gin.Context) {
	permissions, err := h.svc.GetRolePermissions(c.Param("roleCode"))
	if err != nil {
		SendError(c, "500", "获取角色权限失败: "+err.Error())
		return
	}
	SendSuccess(c, permissions)
}

func (h *AdminHandler) HandleUpdateRolePermissions(c *gin.Context) {
	var req model.RolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateRolePermissions(c.Param("roleCode"), req.Permissions); err != nil {
		SendError(c, "500", "更新角色权限失败: "+err.Error())
		return
	}
	SendSuccess(c, nil)
}

// AI Config Handlers

func (h *AdminHandler) HandleListAIProviders(c *gin.Context) {
	list, err := h.svc.ListAIProviders()
	if err != nil {
		SendError(c, "500", "获取 AI 提供商列表失败: "+err.Error())
		return
	}
	SendSuccess(c, list)
}

func (h *AdminHandler) HandleCreateAIProvider(c *gin.Context) {
	var req model.AIProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateAIProvider(req); err != nil {
		SendError(c, "500", "创建 AI 提供商失败: "+err.Error())
		return
	}
	h.aiSvc.ReloadConfig()
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleUpdateAIProvider(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.AIProvider
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateAIProvider(id, req); err != nil {
		SendError(c, "500", "更新 AI 提供商失败: "+err.Error())
		return
	}
	h.aiSvc.ReloadConfig()
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleDeleteAIProvider(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.DeleteAIProvider(id); err != nil {
		SendError(c, "500", "删除 AI 提供商失败: "+err.Error())
		return
	}
	h.aiSvc.ReloadConfig()
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleListAIModels(c *gin.Context) {
	list, err := h.svc.ListAIModels()
	if err != nil {
		SendError(c, "500", "获取 AI 模型列表失败: "+err.Error())
		return
	}
	SendSuccess(c, list)
}

func (h *AdminHandler) HandleCreateAIModel(c *gin.Context) {
	var req model.AIModel
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateAIModel(req); err != nil {
		SendError(c, "500", "创建 AI 模型失败: "+err.Error())
		return
	}
	h.aiSvc.ReloadConfig()
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleUpdateAIModel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.AIModel
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, "400", "请求参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateAIModel(id, req); err != nil {
		SendError(c, "500", "更新 AI 模型失败: "+err.Error())
		return
	}
	h.aiSvc.ReloadConfig()
	SendSuccess(c, nil)
}

func (h *AdminHandler) HandleDeleteAIModel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.DeleteAIModel(id); err != nil {
		SendError(c, "500", "删除 AI 模型失败: "+err.Error())
		return
	}
	h.aiSvc.ReloadConfig()
	SendSuccess(c, nil)
}
