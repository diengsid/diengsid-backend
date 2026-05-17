package middleware

import (
	"github.com/gofiber/fiber/v2"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

func NewAuth(userUseCase *usecase.AuthUseCase) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		token := ctx.Cookies("token")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
		}

		user, err := userUseCase.Verify(ctx.UserContext(), &model.VerifyUserRequest{Token: token})
		if err != nil {
			userUseCase.Log.Warnf("Failed find user by token : %+v", err)
			return err
		}

		ctx.Locals("user", user)
		return ctx.Next()
	}
}

func GetUser(ctx *fiber.Ctx) *model.UserResponse {
	return ctx.Locals("user").(*model.UserResponse)
}

// func GetCompanyId(ctx *fiber.Ctx) string {
// 	user := GetUser(ctx)
// 	return user.CompanyID
// }
