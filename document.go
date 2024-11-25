package onfido

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ------------------------------------------------------------------
//                              DOCUMENT
// ------------------------------------------------------------------

// Document represents a document in the Onfido API
type Document struct {
	ID             string       `json:"id,omitempty"`
	FileType       string       `json:"file_type,omitempty"`
	Type           DocumentType `json:"type,omitempty"`
	Side           string       `json:"side,omitempty"`
	IssuingCountry string       `json:"issuing_country,omitempty"`
	ApplicantID    string       `json:"applicant_id,omitempty"`
	CreatedAt      time.Time    `json:"created_at,omitempty"`
	Href           string       `json:"href,omitempty"`
	DownloadHref   string       `json:"download_href,omitempty"`
	FileName       string       `json:"file_name,omitempty"`
	FileSize       int          `json:"file_size,omitempty"`
}

// DocumentType represents the type of document
//   - The document types declared here are not exhaustive, the API may support more types
type DocumentType string

const (
	DocumentTypeUnknown              DocumentType = "unknown"
	DocumentTypePassport             DocumentType = "passport"
	DocumentTypeDrivingLicence       DocumentType = "driving_licence"
	DocumentTypeNationalIdentityCard DocumentType = "national_identity_card"
	DocumentTypeResidencePermit      DocumentType = "residence_permit"
	DocumentTypeWorkPermit           DocumentType = "work_permit"
	DocumentTypeVoterID              DocumentType = "voter_id"
	DocumentTypeTaxID                DocumentType = "tax_id"
)

type DocumentSide string

const (
	DocumentSideFront DocumentSide = "front"
	DocumentSideBack  DocumentSide = "back"
)

type UploadDocumentPayload struct {
	ApplicantID          string       `json:"applicant_id,omitempty"`
	File                 *os.File     `json:"file,omitempty"`
	FileType             string       `json:"file_type,omitempty"`
	Type                 DocumentType `json:"type,omitempty"`
	Side                 DocumentSide `json:"side,omitempty"`
	IssuingCountry       string       `json:"issuing_country,omitempty"`
	Location             *Location    `json:"location,omitempty"`
	ValidateImageQuality bool         `json:"validate_image_quality,omitempty"`
}

func (ud UploadDocumentPayload) toMultipartMap() (map[string]interface{}, error) {
	file := ud.File

	ud.File = nil
	ub, err := json.Marshal(ud)
	if err != nil {
		return nil, err
	}

	var um map[string]interface{}
	if err := json.Unmarshal(ub, &um); err != nil {
		return nil, err
	}

	um["file"] = file
	return um, nil
}

// ------------------------------------------------------------------
//                              OPTIONS
// ------------------------------------------------------------------

// ------------------------------------------------------------------
//                              METHODS
// ------------------------------------------------------------------

// UploadDocument uploads a document to the Onfido API
func (c *Client) UploadDocument(ctx context.Context, payload UploadDocumentPayload) (*Document, error) {
	var document Document

	req := func() error {
		body, err := c.buildMultipart(payload)
		if err != nil {
			return err
		}

		resp, err := c.client.Post(ctx, "/documents", body, c.getHttpRequestOptions(nil, nil)...)
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &document)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &document, nil
}

// RetrieveDocument retrieves a document from the Onfido API
func (c *Client) RetrieveDocument(ctx context.Context, documentId string) (*Document, error) {
	if documentId == "" {
		return nil, ErrInvalidId
	}

	var document Document

	req := func() error {
		resp, err := c.client.Get(ctx, "/documents/"+documentId, c.getHttpRequestOptions(nil, nil)...)
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &document)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &document, nil
}

// ListDocuments retrieves a list of documents from the Onfido API
func (c *Client) ListDocuments(ctx context.Context, applicantId string) ([]Document, *PageDetails, error) {
	var documents []Document
	var pageDetails PageDetails

	req := func() error {
		params := c.getListDocumentParams(applicantId)
		resp, err := c.client.Get(ctx, "/documents", c.getHttpRequestOptions(params, nil)...)
		if err != nil {
			return err
		}

		var list struct {
			Documents []Document `json:"documents"`
		}
		if err := c.getResponseOrError(resp, &list); err != nil {
			return err
		}

		documents = list.Documents
		pageDetails = c.extractPageDetails(resp.Headers)
		return nil
	}

	if err := c.do(ctx, req); err != nil {
		return nil, nil, err
	}

	return documents, &pageDetails, nil
}

func (c *Client) DownloadDocument(ctx context.Context, documentId string) ([]byte, error) {
	if documentId == "" {
		return nil, ErrInvalidId
	}

	var document []byte

	req := func() error {
		resp, err := c.client.Get(ctx, "/documents/"+documentId+"/download", c.getHttpRequestOptions(nil, nil)...)
		if err != nil {
			return err
		}

		if err := c.getError(resp, true); err != nil {
			return err
		}

		if len(resp.Body) == 0 {
			return fmt.Errorf("unable to download document")
		}

		document = resp.Body

		return nil
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return document, nil
}

func (c *Client) DownloadDocumentNFCFace(ctx context.Context, documentId string) ([]byte, error) {
	if documentId == "" {
		return nil, ErrInvalidId
	}

	var nfcFace []byte

	req := func() error {
		resp, err := c.client.Get(ctx, "/documents/"+documentId+"/nfc_face", c.getHttpRequestOptions(nil, nil)...)
		if err != nil {
			return err
		}

		if err := c.getError(resp, true); err != nil {
			return err
		}

		if len(resp.Body) == 0 {
			return fmt.Errorf("unable to download document")
		}

		nfcFace = resp.Body

		return nil
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return nfcFace, nil
}

func (c *Client) DownloadDocumentVideo(ctx context.Context, documentId string) ([]byte, error) {
	if documentId == "" {
		return nil, ErrInvalidId
	}

	var video []byte

	req := func() error {
		resp, err := c.client.Get(ctx, "/documents/"+documentId+"/video/download", c.getHttpRequestOptions(nil, nil)...)
		if err != nil {
			return err
		}

		if err := c.getError(resp, true); err != nil {
			return err
		}

		if len(resp.Body) == 0 {
			return fmt.Errorf("unable to download document")
		}

		video = resp.Body

		return nil
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return video, nil
}

func (c Client) getListDocumentParams(applicantId string) (params map[string]string) {
	params = map[string]string{
		"applicant_id": applicantId,
	}

	return
}
