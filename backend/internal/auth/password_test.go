package auth

import "testing"

func TestHashAndCompare(t *testing.T) {
	hasher := NewTestHasher()
	hash, err := hasher.Hash("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct-horse" {
		t.Fatal("password stored in clear text")
	}
	if err := hasher.Compare(hash, "correct-horse"); err != nil {
		t.Fatal(err)
	}
	if err := hasher.Compare(hash, "wrong-horse"); err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestCompareMissingHash(t *testing.T) {
	hasher := NewTestHasher()
	if err := hasher.Compare("", "anything"); err == nil {
		t.Fatal("expected mismatch for an empty hash")
	}
}
