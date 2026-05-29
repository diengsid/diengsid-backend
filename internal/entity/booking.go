package entity

import (
	"crypto/rand"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type BookingStatus string
type PaymentStatus string
type FirstPayment string

const (
	StatusPending     BookingStatus = "PENDING"
	StatusWaiting     BookingStatus = "WAITING_PAYMENT"
	StatusUnavailable BookingStatus = "UNAVAILABLE"
	StatusCancelled   BookingStatus = "CANCELLED"
	StatusCheckIn     BookingStatus = "CHECK_IN"
	StatusReview      BookingStatus = "REVIEW"
	StatusDone        BookingStatus = "DONE"

	PaymentUnpaid   PaymentStatus = "UNPAID"
	PaymentPaid     PaymentStatus = "PAID"
	PaymentRefunded PaymentStatus = "REFUNDED"

	FirstPaymentDP   FirstPayment = "DP"
	FirstPaymentFull FirstPayment = "FULL"
)

type Booking struct {
	ID            string        `gorm:"column:id;primaryKey"`
	UserID        string        `gorm:"column:user_id;not null"`
	PropertyID    string        `gorm:"column:property_id;not null"`
	RentableID    string        `gorm:"column:rentable_id;not null"`
	Quantity      int           `gorm:"column:quantity;default:1"`
	GuestCount    int           `gorm:"column:guest_count;default:1"`
	CheckIn       time.Time     `gorm:"column:check_in;not null"`
	CheckOut      time.Time     `gorm:"column:check_out;not null"`
	TotalNight    int           `gorm:"column:total_night;not null"`
	TotalPrice    float64       `gorm:"column:total_price;not null"`
	Discount      float64       `gorm:"column:discount;default:0"`
	Status        BookingStatus `gorm:"column:status;not null;default:PENDING"`
	PaymentStatus PaymentStatus `gorm:"column:payment_status;not null;default:UNPAID"`
	FirstPayment  *FirstPayment `gorm:"column:first_payment"`

	User     User     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Property Property `gorm:"foreignKey:PropertyID;references:ID;constraint:OnDelete:CASCADE"`
	Rentable Rentable `gorm:"foreignKey:RentableID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (Booking) TableName() string {
	return "bookings"
}

const bookingIDChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateBookingID() string {
	suffix := make([]byte, 6)
	for i := range suffix {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(bookingIDChars))))
		suffix[i] = bookingIDChars[n.Int64()]
	}
	return "DID-" + string(suffix)
}

func (b *Booking) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = generateBookingID()
	b.CreatedAt = int64(time.Now().UnixMilli())
	b.UpdatedAt = int64(time.Now().UnixMilli())
	return nil
}
