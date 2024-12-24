package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Location struct {
	Type        string        `bson:"type" json:"type"`
	Coordinates [][][]float64 `bson:"coordinates" json:"coordinates"`
}

type Region struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Province    string             `bson:"province" json:"province"`
	District    string             `bson:"district" json:"district"`
	SubDistrict string             `bson:"sub_district" json:"sub_district"`
	Village     string             `bson:"village" json:"village"`
	Border      Location           `bson:"border" json:"border"`
}

type Roads struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Type       string             `bson:"type" json:"type"`
	Geometry   Geometry           `bson:"geometry" json:"geometry"`
	Properties Properties         `bson:"properties" json:"properties"`
}

type Geometry struct {
	Type        string       `bson:"type" json:"type"`
	Coordinates [][2]float64 `bson:"coordinates" json:"coordinates"`
}

type Properties struct {
	OSMID   int64  `bson:"osm_id" json:"osm_id"`
	Name    string `bson:"name" json:"name"`
	Highway string `bson:"highway" json:"highway"`
}

type LongLat struct {
	Longitude float64 `bson:"long" json:"long"`
	Latitude  float64 `bson:"lat" json:"lat"`
}

type Response struct {
	Response string `json:"response"`
	Info     string `json:"info,omitempty"`
	Status   string `json:"status,omitempty"`
	Location string `json:"location,omitempty"`
}
