package models

import (
	"github.com/rs/zerolog/log"
	"os"
	"time"
	"zeropii/utils"
)

type Customer struct {
	ID         string `json:"id,omitempty" bson:"_id,omitempty"`
	Verified   bool   `json:"verified,omitempty" bson:"verified,omitempty"`
	VerifiedId string `json:"verified_id,omitempty" bson:"verified_id,omitempty"`
	PartnerId  string `json:"partner_id,omitempty" bson:"partner_id,omitempty"`
	Platform   string `json:"platform,omitempty" bson:"platform,omitempty"`
	Consent    bool   `json:"consent,omitempty" bson:"consent,omitempty"`
	//FirstName     string          `json:"first_name,omitempty" bson:"first_name,omitempty"`
	//LastName      string          `json:"last_name,omitempty" bson:"last_name,omitempty"`
	FullName      string          `json:"full_name,omitempty" bson:"full_name,omitempty"`
	Email         string          `json:"email,omitempty" bson:"email,omitempty" pii:"true"`
	Phone         string          `json:"phone,omitempty" bson:"phone,omitempty" pii:"true"`
	DOB           string          `json:"dob,omitempty" bson:"dob,omitempty" pii:"true"`
	MaritalStatus string          `json:"marital_status,omitempty" bson:"marital_status,omitempty"`
	Address       CustomerAddress `json:"address,omitempty" bson:"address,omitempty"`
	Passport      Passport        `json:"passport,omitempty" bson:"passport,omitempty"`
	Pan           Pan             `json:"pan,omitempty" bson:"pan,omitempty"`
	Documents     []Docs          `json:"documents,omitempty" bson:"documents,omitempty"`
	Consents      []ConsentDetail `json:"consents" bson:"consents"`
	CreatedDate   time.Time       `json:"created_date" bson:"created_date"`
	ModifiedDate  time.Time       `json:"modified_date" bson:"modified_date"`
}

type Address struct {
	Street      string `json:"street" bson:"street,omitempty" pii:"true"`
	StreetLine2 string `json:"street_line_2" bson:"streetLine2,omitempty" pii:"true"`
	City        string `json:"city" bson:"city,omitempty" pii:"true"`
	State       string `json:"state" bson:"state,omitempty" pii:"true"`
	Zip         string `json:"zip" bson:"zip,omitempty" pii:"true"`
	Country     string `json:"country" bson:"country" pii:"true"`
}

type CustomerAddress struct {
	CurrentAddress   Address `json:"current_address" bson:"current_address,omitempty"`
	PermanentAddress Address `json:"permanent_address" bson:"permanent_address,omitempty"`
}

type Passport struct {
	PassportNumber       string `json:"passport_number" bson:"passport_number,omitempty" pii:"true"`
	PassportName         string `json:"passport_name" bson:"passport_name,omitempty"`
	PassportIssueDate    string `json:"passport_issue_date" bson:"passport_issue_date,omitempty"`
	PassportExpiryDate   string `json:"passport_expiry_date" bson:"passport_expiry_date,omitempty"`
	PassportDob          string `json:"passport_dob" bson:"passport_dob,omitempty" pii:"true"`
	PassportAddressLine1 string `json:"passport_address_line_1" bson:"passport_address_line_1,omitempty" pii:"true"`
	PassportAddressLine2 string `json:"passport_address_line_2" bson:"passport_address_line_2,omitempty" pii:"true"`
	PassportPostalCode   string `json:"passport_postal_code" bson:"passport_postal_code,omitempty"`
	PassportCity         string `json:"passport_city" bson:"passport_city,omitempty"`
	PassportState        string `json:"passport_state" bson:"passport_state,omitempty"`
	PassportCountry      string `json:"passport_country" bson:"passport_country,omitempty"`
}

type Pan struct {
	PanNumber string `json:"pan_number" bson:"pan_number,omitempty" pii:"true"`
	PanDob    string `json:"pan_dob" bson:"pan_dob,omitempty" pii:"true"`
}

type Docs struct {
	DocType        string `json:"doc_type" bson:"doc_type"`
	DocNumber      string `json:"doc_number" bson:"doc_number" pii:"true"`
	ExpirationDate string `json:"expiration_date" bson:"expiration_date,omitempty"`
	IssuedCountry  string `json:"issued_count" bson:"issued_count,omitempty"`
	ImageUrl       string `json:"image_url" bson:"image_url,omitempty"`
}

type ConsentDetail struct {
	ApplicationName string    `json:"application_name" bson:"application_name"`
	ConsentGiven    bool      `json:"consent_given" bson:"consent_given"`
	ConsentDate     time.Time `json:"consent_date" bson:"consent_date"`
}

// EncryptCustomerPII encrypts all PII fields in the Customer struct
func (c *Customer) EncryptCustomerPII() error {
	var err error

	// Encrypt PII before storing it
	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	encryptedDob, err := utils.Encrypt(c.DOB, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_Dob").
			Msg("Failed to encrypt Dob")
		return err
	}

	encryptedPassportNumber, err := utils.Encrypt(c.Passport.PassportNumber, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_PassportNumber").
			Msg("Failed to encrypt Passport Number")
		return err
	}

	encryptedPassportDob, err := utils.Encrypt(c.Passport.PassportDob, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_PassportDob").
			Msg("Failed to encrypt Passport Dob")
		return err
	}

	encryptedPanNumber, err := utils.Encrypt(c.Pan.PanNumber, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_PanNumber").
			Msg("Failed to encrypt Pan Number")
		return err
	}

	encryptedPanDob, err := utils.Encrypt(c.Pan.PanDob, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_PanDob").
			Msg("Failed to encrypt Pan Dob")

		return err
	}

	c.DOB = encryptedDob
	c.Passport.PassportNumber = encryptedPassportNumber
	c.Passport.PassportDob = encryptedPassportDob
	c.Pan.PanNumber = encryptedPanNumber
	c.Pan.PanDob = encryptedPanDob

	return nil
}

func (c *Customer) DecryptCustomerPII() error {
	var err error

	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	decryptedDob, err := utils.Decrypt(c.DOB, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_Dob").
			Msg("Failed to decrypt Dob")
		return err
	}

	decryptedPassportNumber, err := utils.Decrypt(c.Passport.PassportNumber, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "decrypt_PassportNumber").
			Msg("Failed to decrypt Passport Number")
		return err
	}

	decryptedPassportDob, err := utils.Decrypt(c.Passport.PassportDob, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "encrypt_PassportDob").
			Msg("Failed to decrypt Passport Dob")
		return err
	}

	decryptedPanNumber, err := utils.Decrypt(c.Pan.PanNumber, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "decrypt_PanNumber").
			Msg("Failed to decrypt Pan Number")
		return err
	}

	decryptedPanDob, err := utils.Decrypt(c.Pan.PanDob, encryptionKey)
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "decrypt_PanDob").
			Msg("Failed to decrypt Pan Dob")

		return err
	}

	c.DOB = decryptedDob
	c.Passport.PassportNumber = decryptedPassportNumber
	c.Passport.PassportDob = decryptedPassportDob
	c.Pan.PanNumber = decryptedPanNumber
	c.Pan.PanDob = decryptedPanDob

	return nil
}
