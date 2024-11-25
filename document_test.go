package onfido_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestDocument(t *testing.T) {
	run := setupTestRun(t)
	defer run.teardown()

	applicant, err := run.client.CreateApplicant(run.ctx, onfido.CreateApplicantPayload{
		FirstName: "John",
		LastName:  "DocumentTest",
	})
	if err != nil {
		t.Fatalf("error creating applicant: %v", err)
	}

	file, err := os.Open("./test/medias/license.png")
	if err != nil {
		t.Fatalf("error reading file: %v", err)
	}

	testDocument := &onfido.Document{}

	t.Run("UploadDocument", testUploadDocument(run, applicant.ID, file, testDocument))
	if testDocument.ID == "" {
		t.Fatalf("document ID is empty")
	}
	t.Run("RetrieveDocument", testRetrieveDocument(run, testDocument.ID))
	t.Run("ListDocuments", testListDocuments(run, applicant.ID))
	t.Run("DownloadDocument", testDownloadDocument(run, testDocument.ID))
	t.Run("DownloadDocumentNFCFace", testDownloadDocumentNFCFace(run, testDocument.ID))
	t.Run("DownloadDocumentVideo", testDownloadDocumentVideo(run, testDocument.ID))
}

func testUploadDocument(run *testRun, applicantID string, file *os.File, setDocument *onfido.Document) func(*testing.T) {
	tests := []testCase[onfido.UploadDocumentPayload]{
		{
			name: "UploadWithoutErrors",
			input: onfido.UploadDocumentPayload{
				ApplicantID: applicantID,
				File:        file,
				Type:        onfido.DocumentTypeDrivingLicence,
				FileType:    "png",
				Side:        "front",
			},
		},
	}

	return func(t *testing.T) {
		// sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				document, err := run.client.UploadDocument(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}
				if err != nil {
					t.Fatalf("error uploading document: %v", err)
				}

				// Set the document for later tests
				*setDocument = *document

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, document, "expected document to be uploaded")
				assert.NotEmpty(t, document.ID, "expected document ID to be set")
				assert.NotEmpty(t, document.Href, "expected document href to be set")
				assert.Equal(t, tt.input.Type, document.Type, "expected document type to match")
				assert.Equal(t, tt.input.ApplicantID, document.ApplicantID, "expected applicant ID to match")
				assert.NotNil(t, document.CreatedAt, "expected created at to be set")
			})
		}
	}
}

func testRetrieveDocument(run *testRun, documentId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "RetrieveWithoutErrors",
			input: documentId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "bad_request",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				document, err := run.client.RetrieveDocument(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, document, "expected document to be fetched")
				assert.Equal(t, tt.input, document.ID, "expected document ID to match")
			})
		}
	}
}

func testListDocuments(run *testRun, applicantId string) func(*testing.T) {
	return func(t *testing.T) {
		sleep(t, 5)
		documents, page, err := run.client.ListDocuments(run.ctx, applicantId)
		assert.NoError(t, err, "expected no error listing documents")
		assert.NotNil(t, documents, "expected documents to be fetched")
		assert.NotNil(t, page, "expected page details to be fetched")

		// Verify documents are for the correct applicant
		for _, doc := range documents {
			assert.Equal(t, applicantId, doc.ApplicantID, "expected document to belong to correct applicant")
		}
	}
}

func testDownloadDocument(run *testRun, documentId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "DownloadWithoutErrors",
			input: documentId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "bad_request",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fileBytes, err := run.client.DownloadDocument(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, fileBytes, "expected document content to be downloaded")
				assert.NotEmpty(t, fileBytes, "expected document content to not be empty")

				if os.Getenv("SAVE_FILES") == "true" {
					now := time.Now().Unix()
					saveFile(t, fileBytes, fmt.Sprintf("document-%s-%d.png", tt.input, now))
				}
			})
		}
	}
}

func testDownloadDocumentNFCFace(run *testRun, documentId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "DownloadNFCFaceWithoutErrors",
			input: documentId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		t.Skip("Skipping test")
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				nfcFaceBytes, err := run.client.DownloadDocumentNFCFace(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, nfcFaceBytes, "expected NFC face content to be downloaded")
				assert.NotEmpty(t, nfcFaceBytes, "expected NFC face content to not be empty")
			})
		}
	}
}

func testDownloadDocumentVideo(run *testRun, documentId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "DownloadVideoWithoutErrors",
			input: documentId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		t.Skip("Skipping test")
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				videoBytes, err := run.client.DownloadDocumentVideo(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, videoBytes, "expected video content to be downloaded")
				assert.NotEmpty(t, videoBytes, "expected video content to not be empty")
			})
		}
	}
}

// save to test/medias/debug
func saveFile(t *testing.T, content []byte, filename string) {
	debugDir := filepath.Join("test", "medias", "debug")
	if _, err := os.Stat(debugDir); os.IsNotExist(err) {
		if err := os.MkdirAll(debugDir, 0o755); err != nil {
			t.Fatalf("error creating debug directory: %v", err)
		}
	}

	fullPath := filepath.Join(debugDir, filename)
	file, err := os.Create(fullPath)
	if err != nil {
		t.Fatalf("error creating file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(content); err != nil {
		t.Fatalf("error writing file: %v", err)
	}

	t.Logf("file saved: %s", fullPath)
}
