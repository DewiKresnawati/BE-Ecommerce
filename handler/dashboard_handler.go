package handler

import (
	"be_ecommerce/config"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DashboardData represents the response structure for the dashboard
type DashboardData struct {
	TotalSales   int `json:"totalSales"`
	PendingOrders int `json:"pendingOrders"`
	TotalRevenue int `json:"totalRevenue"`
}

// GetDashboardData retrieves statistics for the seller dashboard
func GetDashboardData(c *fiber.Ctx) error {
	sellerID := c.Query("seller_id")
	if sellerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Seller ID is required"})
	}

	ctx := context.TODO()
	db := config.MongoClient.Database("ecommerce")

	// Get total sales
	ordersCollection := db.Collection("orders")
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"seller_id": sellerID}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "totalSales": bson.M{"$sum": "$total"}}}},
	}
	cursor, err := ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get total sales"})
	}

	var result struct{ TotalSales int }
	if cursor.Next(ctx) {
		cursor.Decode(&result)
	}

	// Get pending orders count
	pendingOrdersCount, err := ordersCollection.CountDocuments(ctx, bson.M{"seller_id": sellerID, "status": "Pending"})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get pending orders"})
	}

	// Get total revenue
	var totalRevenueResult struct{ TotalRevenue int }
	revenuePipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"seller_id": sellerID}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "totalRevenue": bson.M{"$sum": "$total"}}}},
	}
	revenueCursor, err := ordersCollection.Aggregate(ctx, revenuePipeline)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get total revenue"})
	}
	if revenueCursor.Next(ctx) {
		revenueCursor.Decode(&totalRevenueResult)
	}

	// Return response
	dashboardData := DashboardData{
		TotalSales:   result.TotalSales,
		PendingOrders: int(pendingOrdersCount),
		TotalRevenue: totalRevenueResult.TotalRevenue,
	}

	return c.JSON(dashboardData)
}
