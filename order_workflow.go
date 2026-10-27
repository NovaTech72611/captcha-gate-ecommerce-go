package main

import "fmt"

type Signup struct {
	Email          string
	Password       string
	Name           string
	Captcha        string
	WidgetRecordID string
	IP             string
	RequestID      string
}

type Order struct {
	ID       string
	Customer string
	Status   string
	Receipt  string
	Update   string
}

func registerAndCheckout(c *infraiClient, signup Signup, orderID string) (Order, error) {
	passed, err := c.verifyCaptcha(captchaRequest{WidgetRecordID: signup.WidgetRecordID, Token: signup.Captcha, Vendor: "hcaptcha", IP: signup.IP, Action: "signup", ScoreThreshold: 0.5})
	if err != nil {
		return Order{}, err
	}
	if !passed {
		return Order{}, fmt.Errorf("signup rejected: captcha verification did not pass")
	}
	if err := c.createUser(createUserRequest{
		Email: signup.Email, Password: signup.Password, Name: signup.Name,
		Metadata: map[string]any{"order_id": orderID}, Vendor: "hcaptcha",
		Mode: "default_vendor", IdempotencyKey: signup.RequestID,
	}); err != nil {
		return Order{}, err
	}
	return Order{ID: orderID, Customer: signup.Email, Status: "fulfilled", Receipt: "receipt-" + orderID, Update: "Order " + orderID + " is fulfilled"}, nil
}
