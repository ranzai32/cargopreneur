package domain

import "testing"

func TestDeliveryFee(t *testing.T) {
	pickup, ok := DeliveryFee("pickup", 6)
	if !ok || pickup != 0 {
		t.Fatalf("pickup = %v %v", pickup, ok)
	}
	freight, ok := DeliveryFee("freight", 6)
	if !ok || freight != 30.7 {
		t.Fatalf("freight = %v %v", freight, ok)
	}
	express, ok := DeliveryFee("express", 4)
	if !ok || express != 67.6 {
		t.Fatalf("express = %v %v", express, ok)
	}
	if _, ok := DeliveryFee("drone", 1); ok {
		t.Fatal("unknown method accepted")
	}
}
