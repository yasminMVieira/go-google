package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token de autenticação ausente"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de assinatura inválido")
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Não foi possível extrair as claims"})
			c.Abort()
			return
		}

		// Extrair dados do token
		userID, ok := claims["sub"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ID de usuário inválido no token"})
			c.Abort()
			return
		}

		// Armazenar dados do usuário no contexto
		c.Set("userID", userID)
		
		// Extrair roles do token
		if roles, ok := claims["roles"].([]interface{}); ok {
			roleList := make([]string, len(roles))
			for i, role := range roles {
				roleList[i] = role.(string)
			}
			c.Set("roles", roleList)
		}
		
		// Extrair permissões do token
		if permissions, ok := claims["permissions"].([]interface{}); ok {
			permList := make([]string, len(permissions))
			for i, perm := range permissions {
				permList[i] = perm.(string)
			}
			c.Set("permissions", permList)
		}

		c.Next()
	}
}

// RoleMiddleware verifica se o usuário tem um papel específico
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Informações de papel não encontradas"})
			c.Abort()
			return
		}

		roleList, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato inválido para papéis"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roleList {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: papel necessário não encontrado"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionMiddleware verifica se o usuário tem uma permissão específica
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Informações de permissão não encontradas"})
			c.Abort()
			return
		}

		permList, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato inválido para permissões"})
			c.Abort()
			return
		}

		hasPermission := false
		for _, perm := range permList {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: permissão necessária não encontrada"})
			c.Abort()
			return
		}

		c.Next()
	}
}