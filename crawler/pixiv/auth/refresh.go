package auth

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/tidwall/gjson"
)

const (
	CLIENT_ID      = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	CLIENT_SECRET  = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
	AUTH_TOKEN_URL = "https://oauth.secure.pixiv.net/auth/token"
)

func RefreshToken(refresh_token string) (access_token string) {
	c := http.DefaultClient
	payload := new(bytes.Buffer)
	writer := multipart.NewWriter(payload)
	writer.WriteField("client_id", CLIENT_ID)
	writer.WriteField("client_secret", CLIENT_SECRET)
	writer.WriteField("grant_type", "refresh_token")
	writer.WriteField("include_policy", "true")
	writer.WriteField("refresh_token", refresh_token)
	writer.Close()
	req, err := http.NewRequest("POST", AUTH_TOKEN_URL, payload)
	if err != nil {
		log.Println("Refresh token failed: ", err)
		return refresh_token
	}
	req.Header.Add("app-os", "ios")
	req.Header.Add("app-os-version", "14.6")
	req.Header.Add("user-agent", "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)")
	req.Header.Add("Accept-Language", "zh-cn")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return refresh_token
	}
	body, _ := io.ReadAll(resp.Body)
	return gjson.GetBytes(body, "access_token").Str
}
