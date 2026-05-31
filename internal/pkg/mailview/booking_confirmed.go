package mailview

import "fmt"

// BookingConfirmedMailView renders an email for the guest when the host confirms (approved or unavailable).
func BookingConfirmedMailView(name, bookingID, checkIn, checkOut string, totalPrice float64, approved bool) string {
	var title, subtitle, note string
	if approved {
		title = "Booking Anda Dikonfirmasi!"
		subtitle = "Kabar baik! Pemilik properti telah mengonfirmasi ketersediaan. Segera selesaikan pembayaran agar booking terkunci."
		note = "Segera lakukan pembayaran melalui aplikasi sebelum batas waktu habis."
	} else {
		title = "Mohon Maaf, Kamar Tidak Tersedia"
		subtitle = "Pemilik properti menyampaikan bahwa kamar tidak tersedia pada tanggal yang Anda pilih."
		note = "Silakan cari properti lain yang sesuai di Diengs.id."
	}

	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8" /></head>
		<body style="margin:0; background:#f2f2f2; font-family:Arial, sans-serif;">
			<table width="100%%" cellpadding="0" cellspacing="0">
				<tr>
					<td align="center" style="padding:40px 0;">
						<table width="600" style="background:#ffffff; padding:40px; border-radius:16px; border:1px solid #e5e5e5;">

							<tr>
								<td style="padding-bottom:24px;">
									<img src="https://www.image2url.com/r2/default/images/1776225307615-4af4606b-ecc0-476c-a8b9-9f965397be27.png" alt="Diengs.id" width="32" />
								</td>
							</tr>

							<tr>
								<td style="font-size:28px; font-weight:700; color:#222; line-height:1.3;">
									%s
								</td>
							</tr>

							<tr>
								<td style="padding-top:16px; font-size:16px; color:#444; line-height:1.6;">
									Halo <b>%s</b>, %s
								</td>
							</tr>

							<tr>
								<td style="padding-top:28px;">
									<table width="100%%" cellpadding="10" cellspacing="0" style="border:1px solid #e5e5e5; border-radius:8px; font-size:15px; color:#333;">
										<tr style="background:#f9f9f9;">
											<td style="font-weight:600;">ID Booking</td>
											<td>%s</td>
										</tr>
										<tr>
											<td style="font-weight:600;">Check-in</td>
											<td>%s</td>
										</tr>
										<tr style="background:#f9f9f9;">
											<td style="font-weight:600;">Check-out</td>
											<td>%s</td>
										</tr>
										<tr>
											<td style="font-weight:600;">Total Harga</td>
											<td><b>Rp %.0f</b></td>
										</tr>
									</table>
								</td>
							</tr>

							<tr>
								<td style="padding-top:24px; font-size:15px; color:#444; line-height:1.6;">
									%s
								</td>
							</tr>

							<tr><td style="padding:32px 0;"><hr style="border:none; border-top:1px solid #eee;" /></td></tr>

							<tr>
								<td style="padding-bottom:16px;">
									<img src="https://www.image2url.com/r2/default/images/1776225307615-4af4606b-ecc0-476c-a8b9-9f965397be27.png" alt="Diengs.id" width="32" />
								</td>
							</tr>
							<tr>
								<td style="font-size:14px; color:#555; line-height:1.6;">
									Diengsid &mdash; Jawa Tengah, Indonesia
								</td>
							</tr>

						</table>
					</td>
				</tr>
			</table>
		</body>
		</html>
	`, title, name, subtitle, bookingID, checkIn, checkOut, totalPrice, note)
}
