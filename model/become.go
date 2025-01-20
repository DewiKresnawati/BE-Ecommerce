package model

// BecomeSellerRequest represents the payload for the "Become a Seller" request
type BecomeSellerRequest struct {
	StoreName   string `json:"store_name" bson:"store_name" validate:"required"`
	FullAddress string `json:"full_address" bson:"full_address" validate:"required"`
	NIK         string `json:"nik" bson:"nik" validate:"required,len=16,numeric"`
	PhotoSelfie string `json:"photo_selfie" bson:"photo_selfie" validate:"required,base64"`
}

// StoreInfo represents the store information for a seller
type StoreInfo struct {
	StoreName   string `json:"store_name" bson:"store_name"`
	FullAddress string `json:"full_address" bson:"full_address"`
	NIK         string `json:"nik" bson:"nik"`
	PhotoSelfie string `json:"photo_selfie" bson:"photo_selfie"`
}
