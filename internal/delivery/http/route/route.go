package route

import (
	"id.diengs.backend/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App *fiber.App
	// Middleware
	AuthMiddleware   fiber.Handler
	AdminMiddleware  fiber.Handler
	HealthController *http.HealthController
	UploadController *http.UploadController

	AuthController *http.AuthController

	ExperienceController    *http.ExperienceController
	PropertyController      *http.PropertyController
	RentableController      *http.RentableController
	BookingController       *http.BookingController
	AvailabilityController  *http.AvailabilityController
	PaymentController       *http.PaymentController
}

func (c RouteConfig) Setup() {
	c.App.Get("/", c.HealthController.Check)
	c.App.Get("/api/health", c.HealthController.Check)
	c.SetupAuth()
	c.SetupExperience()
	c.SetupProperty()
	c.SetupUpload()
	c.SetupRentable()
	c.SetupBooking()
	c.SetupAvailability()
	c.SetupPayment()
}

func (c RouteConfig) SetupAuth() {
	auth := c.App.Group("/api/auth")
	auth.Post("/register", c.AuthController.Register)
	auth.Post("/login", c.AuthController.Login)
	auth.Post("/send-otp", c.AuthController.SendOtp)
	auth.Post("/verify-otp", c.AuthController.VeriftOtp)
	auth.Post("/google", c.AuthController.AuthGoogle)
	auth.Delete("/_logout", c.AuthController.Logout)

	loggedRoute := auth.Group("/", c.AuthMiddleware)
	loggedRoute.Get("/_current", c.AuthController.Current)
}

func (c RouteConfig) SetupUpload() {
	upload := c.App.Group("/api")
	upload.Post("/upload", c.UploadController.Upload)
	upload.Post("/uploads", c.UploadController.Uploads)

}

func (c RouteConfig) SetupExperience() {
	experience := c.App.Group("/api/experiences")
	experience.Get("/", c.ExperienceController.Search)

	adminRoute := experience.Group("/")
	adminRoute.Post("/", c.ExperienceController.Create)
}

func (c RouteConfig) SetupProperty() {
	property := c.App.Group("/api/properties")
	property.Get("/:id", c.PropertyController.GetByID)

	adminRoute := property.Group("/")
	adminRoute.Post("/", c.PropertyController.Create)
}

func (c RouteConfig) SetupRentable() {
	rentable := c.App.Group("/api/rentables")

	adminRoute := rentable.Group("/")
	adminRoute.Post("/", c.RentableController.Create)
}

func (c RouteConfig) SetupBooking() {
	booking := c.App.Group("/api/bookings", c.AuthMiddleware)
	booking.Post("/", c.BookingController.Create)
	booking.Get("/my", c.BookingController.GetMyBookings)
	booking.Get("/:id", c.BookingController.GetByID)
	booking.Patch("/:id/done", c.BookingController.Complete)

	host := c.App.Group("/api/host", c.AuthMiddleware)
	host.Get("/bookings", c.BookingController.GetHostBookings)
	host.Patch("/bookings/:id/confirm", c.BookingController.ConfirmBooking)
	host.Patch("/bookings/:id/checkout", c.BookingController.Checkout)

	admin := c.App.Group("/api/admin", c.AuthMiddleware, c.AdminMiddleware)
	admin.Get("/bookings", c.BookingController.GetAllBookings)
	admin.Patch("/bookings/:id/confirm", c.BookingController.AdminConfirmBooking)
	admin.Patch("/bookings/:id/checkout", c.BookingController.AdminCheckout)
	admin.Patch("/bookings/:id/done", c.BookingController.AdminComplete)
}

func (c RouteConfig) SetupAvailability() {
	rentable := c.App.Group("/api/rentables")
	rentable.Get("/:id/availability", c.AvailabilityController.Check)
}

func (c RouteConfig) SetupPayment() {
	booking := c.App.Group("/api/bookings", c.AuthMiddleware)
	booking.Get("/:id/payment", c.PaymentController.GetPaymentByBooking)
	booking.Post("/:id/pay", c.PaymentController.CreatePayment)

	c.App.Post("/api/payment/notify", c.PaymentController.HandleNotification)
}
