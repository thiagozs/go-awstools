package awstools

import (
	"testing"
)

func TestNewAWSToolsParams(t *testing.T) {
	region := "us-west-2"
	accessKeyID := "testAccessKey"
	secretKey := "testSecretKey"
	sessionToken := "testSessionToken"
	bufferLimit := 10
	workersRLS := 5

	params, err := newAWSToolsParams(
		WithRegion(region),
		WithAccessKeyID(accessKeyID),
		WithSecretKey(secretKey),
		WithSessionToken(sessionToken),
		WithBufferLimit(bufferLimit),
		WithAmountWorkersRLS(workersRLS),
	)

	if err != nil {
		t.Fatalf("newAWSToolsParams returned error: %v", err)
	}

	if params.Region() != region {
		t.Errorf("Expected region %s, got %s", region, params.Region())
	}

	if params.AccessKeyID() != accessKeyID {
		t.Errorf("Expected accessKeyID %s, got %s", accessKeyID, params.AccessKeyID())
	}

	if params.SecretKey() != secretKey {
		t.Errorf("Expected secretKey %s, got %s", secretKey, params.SecretKey())
	}

	if params.SessionToken() != sessionToken {
		t.Errorf("Expected sessionToken %s, got %s", sessionToken, params.SessionToken())
	}

	if params.BufferLimit() != bufferLimit {
		t.Errorf("Expected bufferLimit %d, got %d", bufferLimit, params.BufferLimit())
	}

	if params.AmountWorkersRLS() != workersRLS {
		t.Errorf("Expected workersRLS %d, got %d", workersRLS, params.AmountWorkersRLS())
	}
}

func TestSetAWSToolsParams(t *testing.T) {
	region := "us-west-2"
	accessKeyID := "testAccessKey"
	secretKey := "testSecretKey"
	sessionToken := "testSessionToken"
	bufferLimit := 10
	workersRLS := 5

	params, err := newAWSToolsParams(
		WithRegion(region),
		WithAccessKeyID(accessKeyID),
		WithSecretKey(secretKey),
		WithSessionToken(sessionToken),
		WithBufferLimit(bufferLimit),
		WithAmountWorkersRLS(workersRLS),
	)

	if err != nil {
		t.Fatalf("newAWSToolsParams returned error: %v", err)
	}

	newRegion := "us-east-1"

	params.SetRegion(newRegion)

	if params.Region() != newRegion {
		t.Errorf("Expected region %s, got %s", newRegion, params.Region())
	}

	newAccessKeyID := "newAccessKeyID"

	params.SetAccessKeyID(newAccessKeyID)

	if params.AccessKeyID() != newAccessKeyID {
		t.Errorf("Expected accessKeyID %s, got %s", newAccessKeyID, params.AccessKeyID())
	}

	newSecretKey := "newSecretKey"

	params.SetSecretKey(newSecretKey)

	if params.SecretKey() != newSecretKey {
		t.Errorf("Expected secretKey %s, got %s", newSecretKey, params.SecretKey())
	}

	newSession := "newSession"

	params.SetSessionToken(newSession)

	if params.SessionToken() != newSession {
		t.Errorf("Expected sessionToken %s, got %s", newSession, params.SessionToken())
	}

	newBufferLimit := 100

	params.SetBufferLimit(newBufferLimit)

	if params.BufferLimit() != newBufferLimit {
		t.Errorf("Expected bufferLimit %d, got %d", newBufferLimit, params.BufferLimit())
	}

	newWorkersRLS := 10

	params.SetAmountWorkersRLS(newWorkersRLS)

	if params.AmountWorkersRLS() != newWorkersRLS {
		t.Errorf("Expected workersRLS %d, got %d", newWorkersRLS, params.AmountWorkersRLS())
	}

}
