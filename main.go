package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	c, err := newInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	email := os.Getenv("INFRAI_DEMO_EMAIL")
	if email == "" {
		log.Fatal("INFRAI_DEMO_EMAIL is required")
	}
	widgetRecordID := os.Getenv("INFRAI_WIDGET_RECORD_ID")
	if widgetRecordID == "" {
		log.Fatal("INFRAI_WIDGET_RECORD_ID is required")
	}
	order, err := registerAndCheckout(c, Signup{
		Email: email, Password: "choose-a-password", Name: "Buyer",
		Captcha: "captcha-token-from-browser", WidgetRecordID: widgetRecordID, IP: "203.0.113.9", RequestID: "signup-order-1001",
	}, "order-1001")
	if err != nil {
		log.Fatal(err)
	}
	result, _ := json.MarshalIndent(order, "", "  ")
	fmt.Println(string(result))
}
