package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterAndCheckoutUsesCaptchaBeforeUserCreate(t *testing.T) {
	called := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/captcha/verify" {
			called = append(called, "captcha")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true,"data":{"passed":true},"error":null,"metadata":{}}`))
			return
		}
		if r.URL.Path == "/v1/auth/user/create" {
			called = append(called, "user")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
			return
		}
		t.Fatalf("unexpected path %s", r.URL.Path)
	}))
	defer server.Close()

	c := &infraiClient{baseURL: server.URL, key: "test-key", http: server.Client(), sleep: func(_ time.Duration) {}}
	order, err := registerAndCheckout(c, Signup{Email: "buyer@example.com", Password: "secret", Name: "Buyer", Captcha: "token", IP: "127.0.0.1", RequestID: "signup-1"}, "order-7")
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != "fulfilled" || order.Receipt != "receipt-order-7" {
		t.Fatalf("unexpected order: %+v", order)
	}
	if len(called) != 2 || called[0] != "captcha" || called[1] != "user" {
		t.Fatalf("business order was %v", called)
	}
}
