package handler

import (
	"be_ecommerce/config"
	"be_ecommerce/model"
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetRoad godoc
// @Summary Mengambil data jalan berdasarkan permintaan
// @Description Mengambil data jalan yang sesuai dengan query yang diberikan berdasarkan geospatial query
// @Accept  json
// @Produce  json
// @Param body body model.RequestBody true "Request Body yang berisi Latitude, Longitude, dan MaxDistance"
// @Success 200 {object} model.Roads "Data jalan yang terdekat dalam format GeoJSON"
// @Failure 400 {object} model.Response "Bad request, body tidak valid"
// @Failure 404 {object} model.Response "Tidak ditemukan jalan terdekat"
// @Failure 500 {object} model.Response "Terjadi kesalahan pada server"
// @Router /api/getroad [post]
func GetRoad(c *fiber.Ctx) error {
	// Mendekodekan body request menjadi RequestBody
	var body RequestBody
	err := json.Unmarshal(c.Body(), &body)
	if err != nil {
		var response model.Response
		response.Status = "Error : Body tidak valid"
		response.Response = err.Error()
		// Mengembalikan response error dengan status 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	// Menghubungkan ke database MongoDB
	Collection := config.MongoClient.Database("petapedia").Collection("roads")
	Ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Geospatial Query Filter
	Filter := bson.M{
		"geometry": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{body.Longitude, body.Latitude},
				},
				"$maxDistance": body.MaxDistance,
			},
		},
	}

	// Melakukan query untuk menemukan road terdekat (menggunakan Find untuk mendapatkan beberapa hasil)
	cursor, err := Collection.Find(Ctx, Filter)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Tidak ditemukan jalan terdekat"
		respn.Response = err.Error()
		// Mengembalikan response error dengan status 404 Not Found
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(respn)
		}
		// Mengembalikan response error lainnya
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}
	defer cursor.Close(Ctx)

	// Membuat slice untuk menampung hasil query
	var roads []model.Roads
	for cursor.Next(Ctx) {
		var road model.Roads
		if err := cursor.Decode(&road); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "Error",
				"message": "Gagal mendekodekan data jalan",
			})
		}
		roads = append(roads, road)
	}

	// Mengembalikan hasil jalan terdekat dalam bentuk JSON
	if len(roads) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "Error",
			"message": "Tidak ada jalan terdekat ditemukan",
		})
	}

	// Membuat struktur GeoJSON yang valid dengan menambahkan "type" dan "features"
	var features []fiber.Map
	for _, road := range roads {
		features = append(features, fiber.Map{
			"type": "Feature", // "type" untuk setiap feature
			"geometry": fiber.Map{
				"type":        road.Geometry.Type,
				"coordinates": road.Geometry.Coordinates,
			},
			"properties": fiber.Map{
				"osm_id":  road.Properties.OSMID,
				"name":    road.Properties.Name,
				"highway": road.Properties.Highway,
			},
		})
	}

	// Menyiapkan response GeoJSON yang valid
	geojsonResponse := fiber.Map{
		"type":     "FeatureCollection", // Menambahkan "type" di tingkat atas
		"features": features,            // Menyertakan array "features" yang berisi data jalan
	}

	// Mengembalikan response dalam format GeoJSON
	return c.JSON(geojsonResponse)
}

// GetRegion godoc
// @Summary Mendapatkan data region berdasarkan koordinat
// @Description Mencari region yang mencakup koordinat latitude dan longitude yang diberikan
// @Accept  json
// @Produce  json
// @Param body body model.LongLat true "Latitude dan Longitude untuk mencari region"
// @Success 200 {object} model.Region "Data region dalam format GeoJSON"
// @Failure 400 {object} model.Response "Bad request, body tidak valid"
// @Failure 404 {object} model.Response "Region tidak ditemukan"
// @Failure 500 {object} model.Response "Terjadi kesalahan pada server"
// @Router /api/getregion [post]
func GetRegion(c *fiber.Ctx) error {
	// Mendekodekan body request menjadi model.LongLat
	var longlat model.LongLat
	err := json.Unmarshal(c.Body(), &longlat)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Body tidak valid"
		respn.Response = err.Error()
		// Mengembalikan response error dengan status 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(respn)
	}

	// Menghubungkan ke database MongoDB
	Collection := config.MongoClient.Database("petapedia").Collection("region")
	Ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Geospatial Query Filter untuk mencari region berdasarkan border
	Filter := bson.M{
		"border": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": bson.M{
					"type":        "Point", // Memastikan jenis geometri adalah "Point"
					"coordinates": []float64{longlat.Longitude, longlat.Latitude},
				},
			},
		},
	}

	// Melakukan query untuk menemukan region yang berdekatan
	var region model.Region
	err = Collection.FindOne(Ctx, Filter).Decode(&region)
	if err != nil {
		var respn model.Response
		respn.Status = "Error : Region tidak ditemukan"
		respn.Response = err.Error()
		// Mengembalikan response error dengan status 404 Not Found
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(respn)
		}
		// Mengembalikan response error lainnya
		return c.Status(fiber.StatusInternalServerError).JSON(respn)
	}

	// Membuat response GeoJSON yang valid
	geojsonResponse := fiber.Map{
		"type": "FeatureCollection", // "type" di tingkat atas
		"features": []fiber.Map{
			{
				"type": "Feature", // "Feature" untuk setiap item dalam features
				"geometry": fiber.Map{
					"type":        region.Border.Type,        // Menggunakan tipe geometri dari field "Border"
					"coordinates": region.Border.Coordinates, // Menggunakan koordinat geometri dari field "Border"
				},
				"properties": fiber.Map{
					"province":     region.Province, // Properti lainnya
					"district":     region.District,
					"sub_district": region.SubDistrict,
					"village":      region.Village,
				},
			},
		},
	}

	// Mengembalikan response dalam format GeoJSON
	return c.Status(fiber.StatusOK).JSON(geojsonResponse)
}
