package fonnte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type ValidateResponse struct {
	Status     bool     `json:"status"`
	Registered []string `json:"registered"`
	Reason     string   `json:"reason"`
}

func ValidateNumber(target string) (bool, error) {
	apiUrl := "https://api.fonnte.com/validate"
	token := "ysYJ1Z8fuEveo6qJsSwZ"

	// 1. Bersihkan karakter aneh tapi biarkan '08' tetap ada (sesuai doc)
	cleanTarget := strings.ReplaceAll(target, " ", "")
	cleanTarget = strings.ReplaceAll(cleanTarget, "-", "")

	// 2. Build Multipart Body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("target", cleanTarget)
	_ = writer.WriteField("countryCode", "62")
	_ = writer.Close()

	// 3. Request
	req, _ := http.NewRequest("POST", apiUrl, body)
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	// DEBUG: Lihat apa yang dibalikin Fonnte
	// Log ini penting buat mastiin Fonnte dapet datanya atau nggak
	fmt.Printf("[FONNTE DOCS MODE] Sent: %s | Status: %v | Registered: %v\n", cleanTarget, result.Status, result.Registered)

	// LOGIC FLEXIBLE: 
	// Asal status true DAN ada isinya di array registered, berarti nomor itu PUNYA WA.
	// Kita nggak peduli Fonnte balikin format 62 atau 08, yang penting DIA TERDAFTAR.
	if result.Status && len(result.Registered) > 0 {
		return true, nil
	}

	return false, nil
}