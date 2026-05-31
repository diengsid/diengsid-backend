package mailview

import "fmt"

// PropertyCreatedMailView renders a notification email for the host when their property is registered.
func PropertyCreatedMailView(hostName, propertyTitle, propertyType, address, propertyURL string) string {
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
									Properti Anda Berhasil Terdaftar!
								</td>
							</tr>

							<tr>
								<td style="padding-top:16px; font-size:16px; color:#444; line-height:1.6;">
									Halo <b>%s</b>, properti Anda telah berhasil terdaftar di Diengs.id dan siap menerima booking dari tamu.
								</td>
							</tr>

							<tr>
								<td style="padding-top:28px;">
									<table width="100%%" cellpadding="10" cellspacing="0" style="border:1px solid #e5e5e5; border-radius:8px; font-size:15px; color:#333;">
										<tr style="background:#f9f9f9;">
											<td style="font-weight:600;">Nama Properti</td>
											<td>%s</td>
										</tr>
										<tr>
											<td style="font-weight:600;">Tipe</td>
											<td>%s</td>
										</tr>
										<tr style="background:#f9f9f9;">
											<td style="font-weight:600;">Alamat</td>
											<td>%s</td>
										</tr>
									</table>
								</td>
							</tr>

							<tr>
								<td style="padding-top:24px; font-size:15px; color:#444; line-height:1.6;">
									Pastikan informasi dan ketersediaan kamar selalu up-to-date agar tamu dapat melakukan booking dengan mudah.
								</td>
							</tr>

							<tr>
								<td style="padding-top:28px; text-align:center;">
									<a href="%s" style="display:inline-block; background:#222; color:#fff; font-size:16px; font-weight:600; padding:14px 32px; border-radius:8px; text-decoration:none;">
										Lihat Properti
									</a>
								</td>
							</tr>

							<tr>
								<td style="padding-top:16px; font-size:13px; color:#888; text-align:center;">
									Atau salin link berikut:<br/>
									<a href="%s" style="color:#444; word-break:break-all;">%s</a>
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
	`, hostName, propertyTitle, propertyType, address, propertyURL, propertyURL, propertyURL)
}
