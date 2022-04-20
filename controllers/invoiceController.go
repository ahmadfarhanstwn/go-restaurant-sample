package controllers

import (
	"context"
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

type InvoiceViewFormat struct {
	Invoice_Id string
	Payment_Method string
	Order_Id string
	Payment_status *string
	Payment_Due interface{}
	Table_Number interface{}
	Payment_Due_Date time.Time
	Order_Details interface{}
}

var InvoiceCollections *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		result, err := InvoiceCollections.Find(context.TODO(), bson.M{})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var allInvoices []bson.M
		if err = result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceId := c.Param("invoice_id")

		var invoice models.Invoice

		err := InvoiceCollections.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_Id)
		invoiceView.Order_Id = invoice.Order_Id
		invoiceView.Payment_Due_Date = invoice.Payment_Due_Date
		
		invoiceView.Payment_Method = "null"
		if invoice.Payment_Method != nil {
			invoiceView.Payment_Method = *invoice.Payment_Method
		}

		invoiceView.Invoice_Id = invoice.Invoice_Id
		invoiceView.Payment_status = *&invoice.Payment_Status
		invoiceView.Payment_Due = allOrderItems[0]["payment_due"]
		invoiceView.Table_Number = allOrderItems[0]["table_number"]
		invoiceView.Order_Details = allOrderItems[0]["order_items"]

		c.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id":invoice.Order_Id}).Decode(&order)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't find the order"})
			return
		}

		status := "PENDING"
		if invoice.Payment_Status != nil {
			invoice.Payment_Status = &status
		}

		invoice.Payment_Due_Date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0,0,1).Format(time.RFC3339))
		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_Id = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := InvoiceCollections.InsertOne(ctx, invoice)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		var invoice models.Invoice
		invoiceId := c.Param("invoice_id")

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"invoiceId" : invoiceId}

		var updateObj primitive.D

		if invoice.Payment_Method != nil {
			updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_Method})
		}

		if invoice.Payment_Status != nil {
			updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_Status})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", invoice.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		status := "PENDING"
		if invoice.Payment_Status == nil {
			invoice.Payment_Status = &status
		}

		result, err := InvoiceCollections.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}