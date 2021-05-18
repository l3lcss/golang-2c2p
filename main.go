package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Userdefined1 struct {
}
type Userdefined2 struct {
}
type Userdefined3 struct {
}
type Userdefined4 struct {
}
type Userdefined5 struct {
}
type Ippperiod struct {
}
type Ippinteresttype struct {
}
type Ippinterestrate struct {
}
type Ippmerchantabsorbrate struct {
}
type Paidchannel struct {
}
type Paidagent struct {
}
type Paymentchannel struct {
}
type Ratequoteid struct {
}
type Originalamount struct {
}

type PaymentResponsePayload struct {
	Timestamp             string                `xml:"timeStamp"`
	Merchantid            string                `xml:"merchantID"`
	Respcode              string                `xml:"respCode"`
	Pan                   string                `xml:"pan"`
	Amt                   string                `xml:"amt"`
	Uniquetransactioncode string                `xml:"uniqueTransactionCode"`
	Tranref               string                `xml:"tranRef"`
	Approvalcode          string                `xml:"approvalCode"`
	Refnumber             string                `xml:"refNumber"`
	Eci                   string                `xml:"eci"`
	Datetime              string                `xml:"dateTime"`
	Status                string                `xml:"status"`
	Failreason            string                `xml:"failReason"`
	Userdefined1          Userdefined1          `xml:"userDefined1"`
	Userdefined2          Userdefined2          `xml:"userDefined2"`
	Userdefined3          Userdefined3          `xml:"userDefined3"`
	Userdefined4          Userdefined4          `xml:"userDefined4"`
	Userdefined5          Userdefined5          `xml:"userDefined5"`
	Ippperiod             Ippperiod             `xml:"ippPeriod"`
	Ippinteresttype       Ippinteresttype       `xml:"ippInterestType"`
	Ippinterestrate       Ippinterestrate       `xml:"ippInterestRate"`
	Ippmerchantabsorbrate Ippmerchantabsorbrate `xml:"ippMerchantAbsorbRate"`
	Paidchannel           Paidchannel           `xml:"paidChannel"`
	Paidagent             Paidagent             `xml:"paidAgent"`
	Paymentchannel        Paymentchannel        `xml:"paymentChannel"`
	Backendinvoice        string                `xml:"backendInvoice"`
	Issuercountry         string                `xml:"issuerCountry"`
	Issuercountrya3       string                `xml:"issuerCountryA3"`
	Bankname              string                `xml:"bankName"`
	Cardtype              string                `xml:"cardType"`
	Processby             string                `xml:"processBy"`
	Paymentscheme         string                `xml:"paymentScheme"`
	Ratequoteid           Ratequoteid           `xml:"rateQuoteID"`
	Originalamount        Originalamount        `xml:"originalAmount"`
	Fxrate                string                `xml:"fxRate"`
	Currencycode          string                `xml:"currencyCode"`
}

const SECURE_PAYMENT_URL = "https://demo2.2c2p.com/2C2PFrontEnd/SecurePayment/Payment.aspx"

//Merchant's account information
const MERCHANT_ID = "JT01"        //Get MerchantID when opening account with 2C2P
const SECRET_KEY = "7jYcp4FxFdf0" //Get SecretKey from 2C2P PGW Dashboard

//Transaction Information
var desc = "2 days 1 night hotel room"
var currencyCode = "702"
var amt = "000000000010"
var panCountry = "SG"

//Customer Information
var cardholderName = "John Doe"

//Request Information
const VERSION = "9.9"

type PaymentRequest struct {
	MerchantID            string `xml:"merchantID"`
	UniqueTransactionCode int64  `xml:"uniqueTransactionCode"`
	Desc                  string `xml:"desc"`
	Amt                   string `xml:"amt"`
	CurrencyCode          string `xml:"currencyCode"`
	PanCountry            string `xml:"panCountry"`
	CardholderName        string `xml:"cardholderName"`
	EncCardData           string `xml:"encCardData"`
}

type PayloadXML struct {
	XMLName   xml.Name `xml:"PaymentRequest"`
	Version   string   `xml:"version"`
	Payload   string   `xml:"payload"`
	Signature string   `xml:"signature"`
}

type PaymentResponse struct {
	Version   string `xml:"version"`
	Payload   string `xml:"payload"`
	Signature string `xml:"signature"`
}

func main() {
	e := echo.New()

	e.POST("/payments", func(c echo.Context) error {
		return paymentHandler(c)
	})

	e.Logger.Fatal(e.Start(":1323"))
}

type PaymentDto struct {
	EncryptedCardInfo string `form:"encryptedCardInfo"`
	MaskedCardInfo    string `form:"maskedCardInfo"`
	ExpMonthCardInfo  string `form:"expMonthCardInfo"`
	ExpYearCardInfo   string `form:"expYearCardInfo"`
}

func paymentHandler(c echo.Context) error {
	payload := &PaymentDto{}

	if err := c.Bind(payload); err != nil {
		return err
	}

	// =-=-=-=-=-=- Construct payment request message  =-=-=-=-=-=-=-=-=-=
	PaymentRequest := &PaymentRequest{
		MerchantID:            MERCHANT_ID,
		UniqueTransactionCode: time.Now().Unix(),
		Desc:                  desc,
		Amt:                   amt,
		CurrencyCode:          currencyCode,
		PanCountry:            panCountry,
		CardholderName:        cardholderName,
		EncCardData:           payload.EncryptedCardInfo,
	}

	payloadOutput, err := xml.Marshal(PaymentRequest)
	if err != nil {
		return err
	}

	paymentPayload := base64.StdEncoding.EncodeToString(payloadOutput)

	h := hmac.New(sha256.New, []byte(SECRET_KEY))
	h.Write([]byte(paymentPayload))
	sha := hex.EncodeToString(h.Sum(nil))

	signature := strings.ToUpper(sha)

	payloadXML := &PayloadXML{
		Version:   VERSION,
		Payload:   paymentPayload,
		Signature: signature,
	}

	payloadXMLOutput, err := xml.Marshal(payloadXML)
	if err != nil {
		return err
	}

	payloadXMLEncoded := base64.StdEncoding.EncodeToString(payloadXMLOutput)
	urlencoded := url.QueryEscape(payloadXMLEncoded)
	requestPayload := "paymentRequest=" + urlencoded
	// =-=-=-=-=-=- Construct payment request message  =-=-=-=-=-=-=-=-=-=

	// =-=-=-=-=-=- Make a request =-=-=-=-=-=-=-=-=-=
	resp, err := http.Post(SECURE_PAYMENT_URL, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(requestPayload)))
	if err != nil {
		return err
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// =-=-=-=-=-=- Make a request =-=-=-=-=-=-=-=-=-=

	// =-=-=-=-=-=- Reading response =-=-=-=-=-=-=-=-=-=
	resBodyDecoded, err := base64.StdEncoding.DecodeString(string(resBody))
	if err != nil {
		return err
	}

	reponsePayLoadXML := &PaymentResponse{
		Version:   VERSION,
		Payload:   paymentPayload,
		Signature: signature,
	}

	if err := xml.Unmarshal(resBodyDecoded, &reponsePayLoadXML); err != nil {
		return err
	}

	paymentResponse, err := base64.StdEncoding.DecodeString(string(reponsePayLoadXML.Payload))
	if err != nil {
		return err
	}

	paymentResponsePayload := &PaymentResponsePayload{}

	if err := xml.Unmarshal(paymentResponse, &paymentResponsePayload); err != nil {
		return err
	}
	// =-=-=-=-=-=- Reading response =-=-=-=-=-=-=-=-=-=

	c.JSON(http.StatusOK, paymentResponsePayload)

	return nil
}
