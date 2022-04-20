package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ahmadfarhanstwn/go-restaurant-management/database"
	"github.com/ahmadfarhanstwn/go-restaurant-management/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")
var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : "error while fetching data from database"})
		}
		var allOrders []bson.M
		if err = result.All(ctx, &allOrders); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var order models.Order
		orderId := c.Param("order_id")

		err := foodCollection.FindOne(ctx, bson.M{"food_id":orderId}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error" : "error occured while fetching data from database"})
		}
		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var order models.Order
		var table models.Table

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return	
		}

		validationErr := validate.Struct(order)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}

		if order.TableID != nil {
			err := tableCollection.FindOne(ctx, bson.M{"tableid":order.TableID}).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("Table was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
				return
			}
		}

		order.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.OrderID = order.ID.Hex()

		result, err := orderCollection.InsertOne(ctx, order)

		if err != nil {
			msg := fmt.Sprintf("error while inserting data to database")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var updateObj primitive.D
		var table models.Table
		var order models.Order
		orderId := c.Param("order_id")
		
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.TableID != nil {
			err := menuCollection.FindOne(
				ctx,
				bson.M{"table_id": order.TableID},
			).Decode(&table)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("menu was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
				return
			}
			updateObj = append(updateObj, bson.E{"menu", order.TableID})
		}

		order.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"update_at", order.Updated_At})

		upsert := true
		filter := bson.M{"order_id": orderId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		result, err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("error while updating data to database")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) string {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	order.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.OrderID = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)
	defer cancel()
	return order.OrderID
}