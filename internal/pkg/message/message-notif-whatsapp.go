package message

import "fmt"

const footer = "\n\nTerima kasih,\nTim Diengs.id"

// ── Booking ───────────────────────────────────────────────────────────────────

func BookingCreatedCustomer(name, bookingID, checkIn, checkOut string, totalNight, guestCount int, totalPrice float64) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nBooking Anda berhasil dibuat dan sedang menunggu konfirmasi pemilik properti."+
			"\n\nDetail Booking:"+
			"\nID           : %s"+
			"\nCheck-in     : %s"+
			"\nCheck-out    : %s"+
			"\nJumlah Malam : %d malam"+
			"\nTamu         : %d tamu"+
			"\nTotal        : Rp %.0f"+
			"\n\nKami akan memberitahu Anda setelah pemilik mengonfirmasi."+
			footer,
		name, bookingID, checkIn, checkOut, totalNight, guestCount, totalPrice,
	)
}

func BookingCreatedHost(hostName, bookingID, propertyTitle, checkIn, checkOut string, totalNight, guestCount int, totalPrice float64) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nAda booking baru masuk untuk properti Anda."+
			"\n\nDetail Booking:"+
			"\nID           : %s"+
			"\nProperti     : %s"+
			"\nCheck-in     : %s"+
			"\nCheck-out    : %s"+
			"\nJumlah Malam : %d malam"+
			"\nTamu         : %d tamu"+
			"\nTotal        : Rp %.0f"+
			"\n\nApakah kamar tersedia pada tanggal tersebut?"+
			"\nBalas pesan ini dengan:"+
			"\n✅ *YA* — jika tersedia"+
			"\n❌ *TIDAK* — jika tidak tersedia"+
			footer,
		hostName, bookingID, propertyTitle, checkIn, checkOut, totalNight, guestCount, totalPrice,
	)
}

func BookingConfirmedCustomer(name, bookingID, checkIn, checkOut string, totalPrice float64) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nKabar baik! Booking Anda telah dikonfirmasi oleh pemilik properti."+
			"\n\nID Booking : %s"+
			"\nCheck-in   : %s"+
			"\nCheck-out  : %s"+
			"\nTotal      : Rp %.0f"+
			"\n\nSegera selesaikan pembayaran agar booking Anda terkonfirmasi penuh."+
			footer,
		name, bookingID, checkIn, checkOut, totalPrice,
	)
}

func BookingUnavailableCustomer(name, bookingID, checkIn, checkOut string) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nMohon maaf, pemilik properti menyatakan bahwa kamar tidak tersedia pada tanggal yang Anda pilih."+
			"\n\nID Booking : %s"+
			"\nCheck-in   : %s"+
			"\nCheck-out  : %s"+
			"\n\nSilakan cari properti lain di Diengs.id."+
			footer,
		name, bookingID, checkIn, checkOut,
	)
}

// ── Payment ───────────────────────────────────────────────────────────────────

func PaymentLinkCustomer(name, bookingID string, totalPrice float64, paymentURL string) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nLink pembayaran untuk booking Anda sudah siap."+
			"\n\nID Booking : %s"+
			"\nTotal      : Rp %.0f"+
			"\n\nSegera selesaikan pembayaran melalui link berikut:\n%s"+
			"\n\nLink berlaku dalam waktu terbatas."+
			footer,
		name, bookingID, totalPrice, paymentURL,
	)
}

func PaymentSuccessCustomer(name, bookingID string, totalPrice float64) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nPembayaran Anda telah berhasil dikonfirmasi."+
			"\n\nID Booking : %s"+
			"\nTotal      : Rp %.0f"+
			"\n\nSelamat menikmati liburan Anda! Tunjukkan konfirmasi ini saat check-in."+
			footer,
		name, bookingID, totalPrice,
	)
}

func PaymentFailedCustomer(name, bookingID string) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nMohon maaf, pembayaran untuk booking Anda gagal diproses."+
			"\n\nID Booking : %s"+
			"\n\nSilakan coba lagi melalui aplikasi."+
			footer,
		name, bookingID,
	)
}

func PaymentExpiredCustomer(name, bookingID string) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nLink pembayaran untuk booking Anda telah kedaluwarsa."+
			"\n\nID Booking : %s"+
			"\n\nSilakan buat link pembayaran baru melalui aplikasi."+
			footer,
		name, bookingID,
	)
}

func PaymentSuccessHost(hostName, bookingID, guestName, propertyTitle, checkIn, checkOut string, guestCount int, totalPrice float64) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nPembayaran telah diterima. Tamu Anda sudah booking dan siap check-in."+
			"\n\nDetail Booking:"+
			"\nID Booking  : %s"+
			"\nNama Tamu   : %s"+
			"\nProperti    : %s"+
			"\nCheck-in    : %s"+
			"\nCheck-out   : %s"+
			"\nJumlah Tamu : %d tamu"+
			"\nTotal       : Rp %.0f"+
			footer,
		hostName, bookingID, guestName, propertyTitle, checkIn, checkOut, guestCount, totalPrice,
	)
}

// ── Property ──────────────────────────────────────────────────────────────────

func PropertyCreatedHost(hostName, propertyTitle, propertyType, address, propertyURL string) string {
	return fmt.Sprintf(
		"Halo %s!"+
			"\n\nTerimakasih telah bergabung di Diengs.id."+
			"\n\nBerikut Detail Properti / Penginapan anda :"+
			"\nNama   : %s"+
			"\nTipe   : %s"+
			"\nAlamat : %s"+
			"\n\nAtau anda bisa lihat Anda di sini:\n%s"+
			footer,
		hostName, propertyTitle, propertyType, address, propertyURL,
	)
}
