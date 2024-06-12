package robokassa

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const FormatISO8601 = "2006-01-02T15:04:05.0000000-07:00"

const CreateInvoiceUrl = "https://auth.robokassa.ru/Merchant/Indexjson.aspx?"

type CreateInvoiceResponse struct {
	InvoiceId    string `json:"invoiceID"`
	MrhInvoiceId int    `json:"mrhInvoiceID"`
	Error        any    `json:"error"`
}

func ApiCreateInvoice(MerchantLogin, Description, Password1 string, OutSum float64, InvId int, ExpirationDate time.Time, IsTest bool) (*CreateInvoiceResponse, error) {
	form := url.Values{}
	form.Add("InvId", fmt.Sprint(InvId))
	form.Add("MerchantLogin", MerchantLogin)
	formattedSum := fmt.Sprintf("%.2f\n", OutSum)
	form.Add("OutSum", formattedSum)
	form.Add("Description", Description)
	form.Add("Culture", "ru")
	form.Add("ExpirationDate", ExpirationDate.Format(FormatISO8601))
	if IsTest {
		form.Add("IsTest", "1")
	}
	//amountString := strconv.FormatFloat(OutSum, 'f', -1, 64)
	sha := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s", MerchantLogin, formattedSum, fmt.Sprint(InvId), Password1)))
	form.Add("SignatureValue", hex.EncodeToString(sha[:]))

	req, err := http.NewRequest("POST", CreateInvoiceUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("fail to create request cause %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to send request cause %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read response body cause %v", err)
	}
	//fmt.Println(string(responseBodyBytes))
	var result CreateInvoiceResponse
	err = json.Unmarshal(responseBodyBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("fail to decode response cause %v", err)
	}
	invoiceId, err := uuid.Parse(result.InvoiceId)
	if err != nil {
		return nil, fmt.Errorf("fail to parse invoice id cause %v", err)
	}
	if invoiceId == uuid.Nil {
		return nil, fmt.Errorf("failed to create an invoice, response: %+v", result)
	}

	return &result, nil
}
